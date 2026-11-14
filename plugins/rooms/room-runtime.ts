import type { MembershipStore, Room, RoomEvent } from "./index.js";
export type RoomRuntime = {
  rooms: Map<string, Room>;
  events: Map<string, RoomEvent[]>;
  sequence: number;
  store: MembershipStore;
  admission?: () => void;
};
export function createRoomRuntime(
  store: MembershipStore = { save: () => {} },
  admission?: () => void,
): RoomRuntime {
  return { rooms: new Map(), events: new Map(), sequence: 0, store, admission };
}
