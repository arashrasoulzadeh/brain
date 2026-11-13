import type { TicTacToeState } from "./index.js";

export function pauseTicTacToeTurn(state: TicTacToeState): TicTacToeState {
  if (state.status !== "active") throw new Error("tic-tac-toe turn is not active");
  return { ...state, board: [...state.board], players: [...state.players] as [string, string], status: "paused", turnDeadlineAt: undefined };
}
