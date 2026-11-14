import { validateProjectId } from "../sdk-typescript/src/index.js";
import type { Invite } from "./types.js";

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
    if (!roomId || !targetPlayerId || ttlSeconds <= 0) throw new Error("invalid invite");
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
