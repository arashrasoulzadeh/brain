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
  const now = Date.now();
  for (const [id, invite] of routeInvites) if (invite.expiresAt <= now) routeInvites.delete(id);
  if (routeInvites.size >= 4096) throw new Error("invite capacity reached");
  const invite = new InviteCreator(routeInvites, randomUUID).create(context.project.id, roomId, playerId, ttl, now);
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
  const candidate = inviteId ? routeInvites.get(inviteId) : [...routeInvites.values()].find(item => item.code === code);
  if (!candidate || candidate.projectId !== context.project.id || candidate.roomId !== roomId)
    throw new Error("invite unavailable");
  const invite = new InviteAccepter(routeInvites).accept(candidate.id, context.project.id, playerId);
  context.plugins.observe.fire("accepted", {});
  return { ...invite, code: "" };
}

export function cancel(context: PluginContext, input: unknown) {
  const payload = requireProjectInput(context, input);
  const roomId = boundedId(payload.roomId, "room id");
  const id = requiredString(payload.inviteId, "invite id is required");
  const invite = routeInvites.get(id);
  if (!invite || invite.projectId !== context.project.id || invite.roomId !== roomId || invite.used || invite.expiresAt <= Date.now())
    throw new Error("invite unavailable");
  routeInvites.delete(id);
  context.plugins.observe.fire("cancelled", {});
  return { id, cancelled: true };
}

/** Operators inspect metadata, never a reusable invitation code. */
export function inspect(context: PluginContext, input: unknown) {
  const payload = requireProjectInput(context, input);
  const roomId = boundedId(payload.roomId, "room id");
  requireOperator(context, "invites.inspect", roomId);
  const now = Date.now();
  return [...routeInvites.values()].filter(invite => invite.projectId === context.project.id && invite.roomId === roomId)
    .map(invite => ({ id: invite.id, roomId: invite.roomId, targetPlayerId: invite.targetPlayerId,
      expiresAt: new Date(invite.expiresAt).toISOString(), status: invite.used ? "accepted" : invite.expiresAt <= now ? "expired" : "pending" }));
}

function boundedId(value: unknown, name: string): string {
  const id = requiredString(value, name + " is required");
  if (id.length > 128) throw new Error("invalid " + name);
  return validateProjectId(id, "invalid " + name);
}
