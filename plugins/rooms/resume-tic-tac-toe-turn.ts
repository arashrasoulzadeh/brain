import type { TicTacToeState } from "./index.js";

export function resumePausedTicTacToeTurn(
  state: TicTacToeState,
  requestedVersion: string,
  pinnedVersion: string,
): TicTacToeState {
  if (state.status !== "paused") throw new Error("tic-tac-toe turn is not paused");
  if (!requestedVersion || requestedVersion !== pinnedVersion)
    throw new Error("pinned room version required");
  return {
    ...state,
    board: [...state.board],
    players: [...state.players] as [string, string],
    status: "active",
  };
}
