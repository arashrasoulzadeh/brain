import { request } from "../sdk-typescript/src/index.js";

export type Invite = { id: string; projectId: string; roomId: string; targetPlayerId: string; expiresAt: number; used: boolean };

export class InviteDirectory {
  private counter = 0;
  private readonly invites = new Map<string, Invite>();
  create(projectId: string, roomId: string, targetPlayerId: string, ttlSeconds: number, now = Date.now()): Invite {
    if (!projectId || !roomId || !targetPlayerId || ttlSeconds <= 0) throw new Error("invalid invite");
    const id = `${projectId}:invite-${++this.counter}`;
    const invite = { id, projectId, roomId, targetPlayerId, expiresAt: now + ttlSeconds * 1000, used: false };
    this.invites.set(id, invite);
    return { ...invite };
  }
}

export function inviteRequest(projectId: string, pluginId: string, version: string, operation: "invites.create" | "invites.accept" | "invites.cancel", roomId: string, payload: unknown = {}) {
  return request(`invites-${operation}-${projectId}-${roomId}`, { protocol: 1, projectId, pluginId, version, operation, payload: { roomId, ...(payload as Record<string, unknown>) } });
}

