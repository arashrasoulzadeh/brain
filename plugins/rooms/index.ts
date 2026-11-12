import { request } from "../sdk-typescript/src/index.js";

export function projectRoomID(projectId: string, roomId: string): string {
  if (!/^[a-zA-Z0-9][a-zA-Z0-9_-]*$/.test(projectId) || !/^[a-zA-Z0-9][a-zA-Z0-9_-]*$/.test(roomId)) throw new Error("invalid project room id");
  return `${projectId}:${roomId}`;
}

export type PlayerSlot = { playerId: string; slot: number };

export type RoomState = "waiting" | "matching" | "active" | "finished" | "closed";

export type Room = { id: string; projectId: string; state: RoomState; maxPlayers: number; slots: PlayerSlot[] };

export class RoomDirectory {
  private readonly rooms = new Map<string, Room>();
  create(projectId: string, roomId: string, maxPlayers: number): Room {
    const id = projectRoomID(projectId, roomId);
    if (maxPlayers < 1 || maxPlayers > 64) throw new Error("invalid max players");
    if (this.rooms.has(id)) throw new Error("room already exists");
    const room: Room = { id, projectId, state: "waiting", maxPlayers, slots: [] };
    this.rooms.set(id, room);
    return { ...room, slots: [] };
  }
  get(projectId: string, roomId: string): Room {
    const room = this.rooms.get(projectRoomID(projectId, roomId));
    if (!room) throw new Error("room not found");
    return { ...room, slots: [...room.slots] };
  }
  join(projectId: string, roomId: string, playerId: string, slot: number): Room {
    const key = projectRoomID(projectId, roomId);
    const room = this.rooms.get(key);
    if (!room) throw new Error("room not found");
    if (room.state !== "waiting" && room.state !== "matching") throw new Error("room not joinable");
    if (!playerId || slot < 0 || slot >= room.maxPlayers || room.slots.some((player) => player.slot === slot || player.playerId === playerId)) throw new Error("invalid membership");
    room.slots.push({ playerId, slot });
    return { ...room, slots: [...room.slots] };
  }
}

export function canTransitionRoom(from: RoomState, to: RoomState): boolean {
  return ({ waiting: ["matching", "closed"], matching: ["active", "waiting", "closed"], active: ["finished", "closed"], finished: ["closed"], closed: [] } as Record<RoomState, RoomState[]>)[from].includes(to);
}

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

