import type { TicTacToeState } from "./index.js";

export function validateTicTacToeMove(
  state: TicTacToeState,
  playerId: string,
  cell: number,
): TicTacToeState {
  if (
    state.status !== "active" ||
    !state.players.includes(playerId) ||
    !Number.isInteger(cell) ||
    cell < 0 ||
    cell >= 9
  )
    throw new Error("invalid tic-tac-toe move");
  if (state.turn !== playerId) throw new Error("out-of-turn move");
  if (state.board[cell] !== null) throw new Error("cell occupied");
  const mark = state.players[0] === playerId ? "X" : "O";
  const board = [...state.board] as ("X" | "O" | null)[];
  board[cell] = mark;
  return { ...state, board, turn: state.players.find((player) => player !== playerId)! };
}
