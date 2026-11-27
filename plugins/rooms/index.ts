export { roomRequest } from "./routes.js";
export { projectRoomID } from "./directory.js";
export type PlayerSlot = { playerId: string; slot: number };
export type RoomState = "waiting" | "matching" | "active" | "finished" | "closed";
export type Room = {
  id: string;
  projectId: string;
  state: RoomState;
  maxPlayers: number;
  slots: PlayerSlot[];
};
export type ConnectionTicket = {
  token: string;
  projectId: string;
  roomId: string;
  playerId: string;
  expiresAt: number;
  used: boolean;
};
export type RoomEvent = {
  type: "player.joined" | "player.left";
  projectId: string;
  roomId: string;
  playerId: string;
  sequence: number;
};
export type MembershipStore = { save(roomId: string, slots: readonly PlayerSlot[]): void };
export { authorizedRoomSnapshot } from "./access.js";
export { authorizeRoomCommand } from "./access.js";
export { ConnectionTicketIssuer } from "./access.js";
export { ConnectionTicketConsumer } from "./access.js";
export { createRoomDirectory } from "./directory.js";
export { canTransitionRoom } from "./directory.js";
export { validateMembership } from "./directory.js";
export { RoomRouteStore, roomRouteStore } from "./routes.js";
export { create, join, reconnect, result, dispatchRoute, manageRoom } from "./routes.js";
