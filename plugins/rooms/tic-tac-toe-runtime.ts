import type { TicTacToeState } from "./index.js";
export type TicTacToeRuntime = { current: TicTacToeState; revision: number };
export function createTicTacToeRuntime(state: TicTacToeState): TicTacToeRuntime {
  return {
    current: { ...state, players: [...state.players] as [string, string], board: [...state.board] },
    revision: 0,
  };
}
