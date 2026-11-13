import { projectRoomID } from "./project-room-id.js";
import type { TicTacToeState } from "./index.js";

export function initializeTicTacToe(projectId: string, roomId: string, firstPlayer: string, secondPlayer: string): TicTacToeState {
  projectRoomID(projectId, roomId);
  if (!firstPlayer.trim() || !secondPlayer.trim() || firstPlayer === secondPlayer) throw new Error("invalid tic-tac-toe players");
  return { projectId, roomId, players: [firstPlayer, secondPlayer], board: Array(9).fill(null), turn: firstPlayer, status: "active" };
}
