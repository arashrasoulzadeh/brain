import { createRoomRuntime } from "./room-runtime.js";
import { RoomCreator } from "./room-creator.js";
import { RoomReader } from "./room-reader.js";
import { RoomMembershipReader } from "./room-membership-reader.js";
import { RoomJoiner } from "./room-joiner.js";
import { RoomLeaver } from "./room-leaver.js";
import { RoomEventReader } from "./room-event-reader.js";
import type { MembershipStore } from "./index.js";
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
  };
}
