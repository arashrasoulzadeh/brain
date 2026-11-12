import { request } from "../sdk-typescript/src/index.js";

export function projectRoomID(projectId: string, roomId: string): string {
  if (!/^[a-zA-Z0-9][a-zA-Z0-9_-]*$/.test(projectId) || !/^[a-zA-Z0-9][a-zA-Z0-9_-]*$/.test(roomId)) throw new Error("invalid project room id");
  return `${projectId}:${roomId}`;
}

export type PlayerSlot = { playerId: string; slot: number };

export function validateMembership(projectId: string, roomId: string, slots: readonly PlayerSlot[], maxPlayers: number): void {
  projectRoomID(projectId, roomId);
  if (maxPlayers < 1 || maxPlayers > 64 || slots.length > maxPlayers) throw new Error("invalid player slots");
  const seen = new Set<number>();
  for (const player of slots) {
    if (!player.playerId || player.slot < 0 || player.slot >= maxPlayers || seen.has(player.slot)) throw new Error("invalid player slots");
    seen.add(player.slot);
  }
}

export function roomRequest(projectId: string, pluginId: string, version: string, operation: "rooms.create" | "rooms.join" | "rooms.reconnect" | "rooms.result", roomId: string, payload: unknown = {}) {
  return request(`rooms-${operation}-${projectId}-${roomId}`, { protocol: 1, projectId, pluginId, version, operation, payload: { roomId, ...(payload as Record<string, unknown>) } });
}

