import { consumeOnce, expiresAt, createResourceRequest, validateProjectId } from "../sdk-typescript/src/index.js";

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
      expiresAt: expiresAt(now, ttlSeconds * 1000, "invalid invite"),
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
    const invite = consumeOnce(this.invites.get(inviteId), now, (invite) =>
      invite.projectId === projectId && invite.targetPlayerId === playerId,
    );
    if (!invite) throw new Error("invite unavailable");
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
