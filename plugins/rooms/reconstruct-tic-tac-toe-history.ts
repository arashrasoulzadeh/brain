import { validateTicTacToeMove } from "./validate-tic-tac-toe-move.js";
import type { CommittedTicTacToeMove, TicTacToeState } from "./index.js";
export function reconstructTicTacToeHistory(
  initialState: TicTacToeState,
  committedMoves: readonly CommittedTicTacToeMove[],
): TicTacToeState {
  return committedMoves.reduce((state, outcome, expectedRevision) => {
    if (outcome.revision !== expectedRevision || !outcome.playerId.trim())
      throw new Error("invalid committed game history");
    try {
      return validateTicTacToeMove(state, outcome.playerId, outcome.cell);
    } catch {
      throw new Error("invalid committed game history");
    }
  }, initialState);
}
