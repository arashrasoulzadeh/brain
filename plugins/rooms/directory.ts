import { validateProjectId } from "../sdk-typescript/src/index.js";
import { type MembershipStore, type Room, type RoomEvent, type RoomState, type PlayerSlot } from "./index.js";

export function projectRoomID(projectId: string, roomId: string): string {
  validateProjectId(projectId);
  if (!/^[a-zA-Z0-9][a-zA-Z0-9_-]*$/.test(roomId)) throw new Error("invalid project room id");
  return `${projectId}:${roomId}`;
}

type RoomRuntime = {
  rooms: Map<string, Room>;
  events: Map<string, RoomEvent[]>;
  sequence: number;
  drainUntil?: number;
  store: MembershipStore;
  admission?: () => void;
};
function createRoomRuntime(
  store: MembershipStore = { save: () => {} },
  admission?: () => void,
): RoomRuntime {
  return { rooms: new Map(), events: new Map(), sequence: 0, store, admission };
}

function isRoomDraining(runtime: RoomRuntime): boolean {
  return runtime.drainUntil !== undefined && runtime.drainUntil > Date.now();
}

export function canTransitionRoom(from: RoomState, to: RoomState): boolean {
  return (
    {
      waiting: ["matching", "closed"],
      matching: ["active", "waiting", "closed"],
      active: ["finished", "closed"],
      finished: ["closed"],
      closed: [],
    } as Record<RoomState, RoomState[]>
  )[from].includes(to);
}

export function validateMembership(
  projectId: string,
  roomId: string,
  slots: readonly PlayerSlot[],
  maxPlayers: number,
): void {
  projectRoomID(projectId, roomId);
  if (maxPlayers < 1 || maxPlayers > 64 || slots.length > maxPlayers)
    throw new Error("invalid player slots");
  const seen = new Set<number>();
  for (const player of slots) {
    if (!player.playerId || player.slot < 0 || player.slot >= maxPlayers || seen.has(player.slot))
      throw new Error("invalid player slots");
    seen.add(player.slot);
  }
}

function startRoomDrain(runtime: RoomRuntime, deadline: number): void {
  if (!Number.isFinite(deadline) || deadline <= Date.now())
    throw new Error("invalid drain deadline");
  runtime.drainUntil = deadline;
}

function stopRoomDrain(runtime: RoomRuntime): void {
  runtime.drainUntil = undefined;
}

export function createRoomDirectory(store?: MembershipStore, admission?: () => void) {
  const runtime = createRoomRuntime(store, admission);

  function create(projectId: string, roomId: string, maxPlayers: number): Room {
    runtime.admission?.();
    if (isRoomDraining(runtime)) throw new Error("room admissions disabled during drain");
    const id = projectRoomID(projectId, roomId);
    if (maxPlayers < 1 || maxPlayers > 64) throw new Error("invalid max players");
    if (runtime.rooms.has(id)) throw new Error("room already exists");
    const room: Room = { id, projectId, state: "waiting", maxPlayers, slots: [] };
    runtime.rooms.set(id, room);
    runtime.events.set(id, []);
    runtime.store.save(id, []);
    return { ...room, slots: [] };
  }

  function get(projectId: string, roomId: string) {
    const room = runtime.rooms.get(projectRoomID(projectId, roomId));
    if (!room) throw new Error("room not found");
    return { ...room, slots: [...room.slots] };
  }

  function isMember(projectId: string, roomId: string, playerId: string): boolean {
    return (
      runtime.rooms
        .get(projectRoomID(projectId, roomId))
        ?.slots.some((player) => player.playerId === playerId) ?? false
    );
  }

  function join(projectId: string, roomId: string, playerId: string, slot: number) {
    const key = projectRoomID(projectId, roomId);
    const room = runtime.rooms.get(key);
    if (!room) throw new Error("room not found");
    if (room.state !== "waiting" && room.state !== "matching") throw new Error("room not joinable");
    if (
      !playerId ||
      slot < 0 ||
      slot >= room.maxPlayers ||
      room.slots.some((player) => player.slot === slot || player.playerId === playerId)
    )
      throw new Error("invalid membership");
    room.slots.push({ playerId, slot });
    runtime.store.save(key, room.slots);
    runtime.events.get(key)?.push({
      type: "player.joined",
      projectId,
      roomId,
      playerId,
      sequence: ++runtime.sequence,
    });
    return { ...room, slots: [...room.slots] };
  }

  function leave(projectId: string, roomId: string, playerId: string): void {
    const key = projectRoomID(projectId, roomId);
    const room = runtime.rooms.get(key);
    if (!room || !room.slots.some((player) => player.playerId === playerId))
      throw new Error("player not in room");
    room.slots = room.slots.filter((player) => player.playerId !== playerId);
    runtime.store.save(key, room.slots);
    runtime.events.get(key)?.push({
      type: "player.left",
      projectId,
      roomId,
      playerId,
      sequence: ++runtime.sequence,
    });
  }

  function events(projectId: string, roomId: string) {
    return [...(runtime.events.get(projectRoomID(projectId, roomId)) ?? [])];
  }

  return {
    create, get, isMember, join, leave, events,
    reconnect(projectId: string, roomId: string, playerId: string) {
      const room = get(projectId, roomId);
      if (!room.slots.some((slot) => slot.playerId === playerId))
        throw new Error("player not in room");
      return room;
    },
    eventsSince(projectId: string, roomId: string, sequence: number) {
      get(projectId, roomId);
      if (!Number.isInteger(sequence) || sequence < 0) throw new Error("invalid event sequence");
      return events(projectId, roomId).filter((event) => event.sequence > sequence);
    },
    startDrain: (deadline: number) => startRoomDrain(runtime, deadline),
    stopDrain: () => stopRoomDrain(runtime),
  };
}
