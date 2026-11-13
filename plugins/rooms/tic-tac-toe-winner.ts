import type { TicTacToeState } from "./index.js";

export function ticTacToeWinner(state: TicTacToeState): string | undefined {
  const lines = [[0, 1, 2], [3, 4, 5], [6, 7, 8], [0, 3, 6], [1, 4, 7], [2, 5, 8], [0, 4, 8], [2, 4, 6]];
  for (const [a, b, c] of lines) if (state.board[a] && state.board[a] === state.board[b] && state.board[a] === state.board[c]) return state.players[state.board[a] === "X" ? 0 : 1];
  return undefined;
}
