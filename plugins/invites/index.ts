import { pluginRequest, validateProjectId } from "../sdk-typescript/src/index.js";

export type Invite = {
  id: string;
  code: string;
  projectId: string;
  roomId: string;
  targetPlayerId: string;
  expiresAt: number;
  used: boolean;
};

export class InviteCreator {
  constructor(
    private readonly invites: Map<string, Invite>,
    private readonly nextId: () => number,
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

export function inviteRequest(
  projectId: string,
  pluginId: string,
  version: string,
  operation: "invites.create" | "invites.accept" | "invites.cancel",
  roomId: string,
  payload: unknown = {},
) {
  return pluginRequest(
    "invites",
    projectId,
    pluginId,
    version,
    operation,
    { roomId, ...(payload as Record<string, unknown>) },
    roomId,
  );
}
