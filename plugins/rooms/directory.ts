import { projectRoomID } from "../sdk-typescript/src/index.js";
import { type MembershipStore, type Room, type RoomEvent, type RoomState, type PlayerSlot } from "./index.js";

export { projectRoomID } from "../sdk-typescript/src/index.js";

export type RoomRuntime = {
  rooms: Map<string, Room>;
  events: Map<string, RoomEvent[]>;
  sequence: number;
  drainUntil?: number;
  store: MembershipStore;
  admission?: () => void;
};
export function createRoomRuntime(
  store: MembershipStore = { save: () => {} },
  admission?: () => void,
): RoomRuntime {
  return { rooms: new Map(), events: new Map(), sequence: 0, store, admission };
}

export function isRoomDraining(runtime: RoomRuntime): boolean {
  return runtime.drainUntil !== undefined && runtime.drainUntil > Date.now();
}

export class RoomCreator {
  constructor(private readonly runtime: RoomRuntime) {}
  create(projectId: string, roomId: string, maxPlayers: number): Room {
    this.runtime.admission?.();
    if (isRoomDraining(this.runtime)) throw new Error("room admissions disabled during drain");
    const id = projectRoomID(projectId, roomId);
    if (maxPlayers < 1 || maxPlayers > 64) throw new Error("invalid max players");
    if (this.runtime.rooms.has(id)) throw new Error("room already exists");
    const room: Room = { id, projectId, state: "waiting", maxPlayers, slots: [] };
    this.runtime.rooms.set(id, room);
    this.runtime.events.set(id, []);
    this.runtime.store.save(id, []);
    return { ...room, slots: [] };
  }
}

export class RoomReader {
  constructor(private readonly runtime: RoomRuntime) {}
  get(projectId: string, roomId: string) {
    const room = this.runtime.rooms.get(projectRoomID(projectId, roomId));
    if (!room) throw new Error("room not found");
    return { ...room, slots: [...room.slots] };
  }
}

export class RoomMembershipReader {
  constructor(private readonly runtime: RoomRuntime) {}
  isMember(projectId: string, roomId: string, playerId: string): boolean {
    return (
      this.runtime.rooms
        .get(projectRoomID(projectId, roomId))
        ?.slots.some((player) => player.playerId === playerId) ?? false
    );
  }
}

export class RoomJoiner {
  constructor(private readonly runtime: RoomRuntime) {}
  join(projectId: string, roomId: string, playerId: string, slot: number) {
    const key = projectRoomID(projectId, roomId);
    const room = this.runtime.rooms.get(key);
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
    this.runtime.store.save(key, room.slots);
    this.runtime.events.get(key)?.push({
      type: "player.joined",
      projectId,
      roomId,
      playerId,
      sequence: ++this.runtime.sequence,
    });
    return { ...room, slots: [...room.slots] };
  }
}

export class RoomLeaver {
  constructor(private readonly runtime: RoomRuntime) {}
  leave(projectId: string, roomId: string, playerId: string): void {
    const key = projectRoomID(projectId, roomId);
    const room = this.runtime.rooms.get(key);
    if (!room || !room.slots.some((player) => player.playerId === playerId))
      throw new Error("player not in room");
    room.slots = room.slots.filter((player) => player.playerId !== playerId);
    this.runtime.store.save(key, room.slots);
    this.runtime.events.get(key)?.push({
      type: "player.left",
      projectId,
      roomId,
      playerId,
      sequence: ++this.runtime.sequence,
    });
  }
}

export class RoomEventReader {
  constructor(private readonly runtime: RoomRuntime) {}
  events(projectId: string, roomId: string) {
    return [...(this.runtime.events.get(projectRoomID(projectId, roomId)) ?? [])];
  }
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

export function startRoomDrain(runtime: RoomRuntime, deadline: number): void {
  if (!Number.isFinite(deadline) || deadline <= Date.now())
    throw new Error("invalid drain deadline");
  runtime.drainUntil = deadline;
}

export function stopRoomDrain(runtime: RoomRuntime): void {
  runtime.drainUntil = undefined;
}

export function createRoomDirectory(store?: MembershipStore, admission?: () => void) {
  const runtime = createRoomRuntime(store, admission);
  const creator = new RoomCreator(runtime);
  const reader = new RoomReader(runtime);
  const membership = new RoomMembershipReader(runtime);
  const joiner = new RoomJoiner(runtime);
  const leaver = new RoomLeaver(runtime);
  const events = new RoomEventReader(runtime);
  return {
    create: creator.create.bind(creator),
    get: reader.get.bind(reader),
    isMember: membership.isMember.bind(membership),
    reconnect: (projectId: string, roomId: string, playerId: string) => {
      const room = reader.get(projectId, roomId);
      if (!room.slots.some((slot) => slot.playerId === playerId))
        throw new Error("player not in room");
      return room;
    },
    join: joiner.join.bind(joiner),
    leave: leaver.leave.bind(leaver),
    events: events.events.bind(events),
    eventsSince: (projectId: string, roomId: string, sequence: number) => {
      reader.get(projectId, roomId);
      if (!Number.isInteger(sequence) || sequence < 0) throw new Error("invalid event sequence");
      return events.events(projectId, roomId).filter((event) => event.sequence > sequence);
    },
    startDrain: (deadline: number) => startRoomDrain(runtime, deadline),
    stopDrain: () => stopRoomDrain(runtime),
  };
}
