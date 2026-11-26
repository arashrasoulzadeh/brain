export {
  roomRequest, RoomRouteStore, roomRouteStore, create, join, reconnect, result,
  dispatchRoute,
} from "./routes.js";
export {
  projectRoomID, createRoomDirectory, canTransitionRoom, validateMembership,
} from "./directory.js";
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
export type TicTacToeState = {
  projectId: string;
  roomId: string;
  players: readonly [string, string];
  board: readonly ("X" | "O" | null)[];
  turn: string;
  status: "active" | "paused";
  turnDeadlineAt?: number;
};
export type CommittedTicTacToeMove = { revision: number; playerId: string; cell: number };
export {
  authorizedRoomSnapshot, authorizeRoomCommand, ConnectionTicketIssuer,
  ConnectionTicketConsumer, authorizedTicTacToeState, authorizeTicTacToeCommand,
  CommandDeduplicator, validateTicTacToeCommandPayload,
} from "./access.js";
export {
  initializeTicTacToe, validateTicTacToeMove, ticTacToeWinner, ticTacToeDraw,
  validateTicTacToeRuleResult, TIC_TAC_TOE_RULES_SOURCE, publishTicTacToeRules,
  createTicTacToeCoordinator, TicTacToeRevisionReader, TicTacToeMover,
} from "./gameplay.js";
export {
  expireTicTacToeTurn, pauseTicTacToeTurn, resumePausedTicTacToeTurn,
  reconstructTicTacToeHistory, serializeTicTacToeState, reconnectTicTacToe,
  hiddenTicTacToeView,
} from "./recovery.js";
