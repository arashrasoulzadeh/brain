import { pluginRequest } from "../sdk-typescript/src/index.js";

export type Invite = { id: string; code: string; projectId: string; roomId: string; targetPlayerId: string; expiresAt: number; used: boolean };

export class InviteDirectory {
  private counter = 0;
  private readonly invites = new Map<string, Invite>();
  create(projectId: string, roomId: string, targetPlayerId: string, ttlSeconds: number, now = Date.now()): Invite {
    if (!projectId || !roomId || !targetPlayerId || ttlSeconds <= 0) throw new Error("invalid invite");
    const id = `${projectId}:invite-${++this.counter}`;
    const invite = { id, code: `INVITE-${projectId}-${++this.counter}`, projectId, roomId, targetPlayerId, expiresAt: now + ttlSeconds * 1000, used: false };
    this.invites.set(id, invite);
    return { ...invite };
  }

  accept(inviteId: string, projectId: string, playerId: string, now = Date.now()): Invite {
    const invite = this.invites.get(inviteId);
    if (!invite || invite.projectId !== projectId || invite.targetPlayerId !== playerId || invite.used || invite.expiresAt <= now) throw new Error("invite unavailable");
    invite.used = true;
    return { ...invite };
  }

  acceptCode(code: string, projectId: string, playerId: string, now = Date.now()): Invite {
    const invite = [...this.invites.values()].find((candidate) => candidate.code === code);
    if (!invite) throw new Error("invite unavailable");
    return this.accept(invite.id, projectId, playerId, now);
  }
}

export function inviteRequest(projectId: string, pluginId: string, version: string, operation: "invites.create" | "invites.accept" | "invites.cancel", roomId: string, payload: unknown = {}) {
  return pluginRequest("invites", projectId, pluginId, version, operation, { roomId, ...(payload as Record<string, unknown>) }, roomId);
}

