import type { TicTacToeState } from "./index.js";

export function expireTicTacToeTurn(
  state: TicTacToeState,
  now: number,
  nextDeadlineAt: number,
): TicTacToeState {
  if (
    state.status !== "active" ||
    !Number.isFinite(now) ||
    !Number.isFinite(nextDeadlineAt) ||
    nextDeadlineAt <= now
  )
    throw new Error("invalid tic-tac-toe timeout");
  if (state.turnDeadlineAt === undefined || now < state.turnDeadlineAt)
    throw new Error("tic-tac-toe turn not expired");
  return {
    ...state,
    turn: state.players.find((player) => player !== state.turn)!,
    turnDeadlineAt: nextDeadlineAt,
  };
}
