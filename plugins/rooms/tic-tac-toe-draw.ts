import { ticTacToeWinner } from "./tic-tac-toe-winner.js";
import type { TicTacToeState } from "./index.js";

export function ticTacToeDraw(state: TicTacToeState): boolean {
  return !ticTacToeWinner(state) && state.board.every((cell) => cell !== null);
}
