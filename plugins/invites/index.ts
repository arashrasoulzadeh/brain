import { randomUUID } from "node:crypto";
import { createResourceRequest, requireProjectInput, requireOperator, requiredString, validateProjectId, type PluginContext } from "../sdk-typescript/src/index.js";

export type Invite = {
  id: string;
  code: string;
  projectId: string;
  roomId: string;
  targetPlayerId: string;
  expiresAt: number;
  used: boolean;
};

const routeInvites = new Map<string, Invite>();

type InviteSnapshot = { records?: Invite[]; idPart?: string; codePart?: string; now?: number };
function storedInvites(payload: Record<string, unknown>): Map<string, Invite> {
  if (!Object.hasOwn(payload, "_brainInvites")) return routeInvites;
  const snapshot = payload._brainInvites as InviteSnapshot;
  if (!snapshot || !Array.isArray(snapshot.records)) throw new Error("invalid stored invitations");
  return new Map(snapshot.records.map(record => [record.id, record]));
}

export class InviteCreator {
  constructor(
    private readonly invites: Map<string, Invite>,
    private readonly nextId: () => string | number,
  ) {}
  create(
    projectId: string,
    roomId: string,
    targetPlayerId: string,
    ttlSeconds: number,
    now = Date.now(),
  ): Invite {
    validateProjectId(projectId, "invalid invite");
    if (
      !roomId ||
      !targetPlayerId ||
      !Number.isFinite(now) ||
      !Number.isFinite(ttlSeconds) ||
      ttlSeconds <= 0
    )
      throw new Error("invalid invite");
    const id = `${projectId}:invite-${this.nextId()}`;
    const invite = {
      id,
      code: `INVITE-${projectId}-${this.nextId()}`,
      projectId,
      roomId,
      targetPlayerId,
      expiresAt: now + ttlSeconds * 1000,
      used: false,
    };
    this.invites.set(id, invite);
    return { ...invite };
  }
}

export class InviteAccepter {
  constructor(private readonly invites: Map<string, Invite>) {}
  accept(inviteId: string, projectId: string, playerId: string, now = Date.now()): Invite {
    validateProjectId(projectId, "invite unavailable");
    if (!inviteId || !playerId || !Number.isFinite(now)) throw new Error("invite unavailable");
    const invite = this.invites.get(inviteId);
    if (
      !invite ||
      invite.projectId !== projectId ||
      invite.targetPlayerId !== playerId ||
      invite.used ||
      invite.expiresAt <= now
    )
      throw new Error("invite unavailable");
    invite.used = true;
    return { ...invite };
  }
}

export class InviteCodeAccepter {
  private readonly accepter: InviteAccepter;
  constructor(private readonly invites: Map<string, Invite>) {
    this.accepter = new InviteAccepter(invites);
  }
  acceptCode(code: string, projectId: string, playerId: string, now = Date.now()): Invite {
    const invite = [...this.invites.values()].find((candidate) => candidate.code === code);
    if (!invite) throw new Error("invite unavailable");
    return this.accepter.accept(invite.id, projectId, playerId, now);
  }
}

export const inviteRequest = createResourceRequest<"invites.create" | "invites.accept" | "invites.cancel">("invites", "roomId");

/** Operator worker routes. The host must authorize the target player before exposing these to clients. */
export function create(context: PluginContext, input: unknown): Invite {
  const payload = requireProjectInput(context, input);
  const roomId = boundedId(payload.roomId, "room id");
  const playerId = boundedId(payload.targetPlayerId ?? payload.target, "target player id");
  const ttl = payload.ttlSeconds ?? 60;
  if (typeof ttl !== "number" || !Number.isSafeInteger(ttl) || ttl < 1 || ttl > 86400)
    throw new Error("invalid invite ttl");
  const stored = payload._brainInvites as InviteSnapshot | undefined;
  const now = stored?.now ?? Date.now();
  const invites = stored ? new Map<string, Invite>() : routeInvites;
  for (const [id, invite] of invites) if (invite.expiresAt <= now) invites.delete(id);
  if (invites.size >= 4096) throw new Error("invite capacity reached");
  let nextId = randomUUID as () => string;
  if (stored) {
    if (!stored.idPart || !stored.codePart || !Number.isSafeInteger(stored.now)) throw new Error("invalid stored invitation allocation");
    const parts = [stored.idPart, stored.codePart];
    nextId = () => parts.shift()!;
  }
  const invite = new InviteCreator(invites, nextId).create(context.project.id, roomId, playerId, ttl, now);
  context.plugins.observe.fire("created", {});
  return invite;
}

export function accept(context: PluginContext, input: unknown): Invite {
  const payload = requireProjectInput(context, input);
  const roomId = boundedId(payload.roomId, "room id");
  const playerId = boundedId(payload.playerId, "player id");
  const inviteId = typeof payload.inviteId === "string" ? payload.inviteId : "";
  const code = typeof payload.code === "string" ? payload.code : "";
  if (Boolean(inviteId) === Boolean(code) || inviteId.length > 256 || code.length > 256)
    throw new Error("invite unavailable");
  const invites = storedInvites(payload);
  const candidate = inviteId ? invites.get(inviteId) : [...invites.values()].find(item => item.code === code);
  if (!candidate || candidate.projectId !== context.project.id || candidate.roomId !== roomId)
    throw new Error("invite unavailable");
  const invite = new InviteAccepter(invites).accept(candidate.id, context.project.id, playerId);
  context.plugins.observe.fire("accepted", {});
  return { ...invite, code: "" };
}

export function cancel(context: PluginContext, input: unknown) {
  const payload = requireProjectInput(context, input);
  const roomId = boundedId(payload.roomId, "room id");
  const id = requiredString(payload.inviteId, "invite id is required");
  const invites = storedInvites(payload);
  const invite = invites.get(id);
  if (!invite || invite.projectId !== context.project.id || invite.roomId !== roomId || invite.used || invite.expiresAt <= Date.now())
    throw new Error("invite unavailable");
  invites.delete(id);
  context.plugins.observe.fire("cancelled", {});
  return { id, cancelled: true };
}

/** Operators inspect metadata, never a reusable invitation code. */
export function inspect(context: PluginContext, input: unknown) {
  const payload = requireProjectInput(context, input);
  const roomId = boundedId(payload.roomId, "room id");
  requireOperator(context, "invites.inspect", roomId);
  const now = Date.now();
  return [...storedInvites(payload).values()].filter(invite => invite.projectId === context.project.id && invite.roomId === roomId)
    .map(invite => ({ id: invite.id, roomId: invite.roomId, targetPlayerId: invite.targetPlayerId,
      expiresAt: new Date(invite.expiresAt).toISOString(), status: invite.used ? "accepted" : invite.expiresAt <= now ? "expired" : "pending" }));
}

function boundedId(value: unknown, name: string): string {
  const id = requiredString(value, name + " is required");
  if (id.length > 128) throw new Error("invalid " + name);
  return validateProjectId(id, "invalid " + name);
}
