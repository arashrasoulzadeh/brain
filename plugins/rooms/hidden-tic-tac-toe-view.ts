import type { TicTacToeState } from "./index.js";
export function hiddenTicTacToeView<T>(
  state: TicTacToeState,
  playerId: string,
  privateState: Readonly<Record<string, T>>,
): TicTacToeState & { privateState?: T } {
  const view = {
    ...state,
    players: [...state.players] as [string, string],
    board: [...state.board],
  };
  return {
    ...view,
    privateState: state.players.includes(playerId) ? privateState[playerId] : undefined,
  };
}
