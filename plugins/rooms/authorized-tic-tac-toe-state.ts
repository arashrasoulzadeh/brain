import type { TicTacToeState } from "./index.js";
export function authorizedTicTacToeState(
  state: TicTacToeState,
  projectId: string,
  playerId: string,
): TicTacToeState {
  if (state.projectId !== projectId || !state.players.includes(playerId))
    throw new Error("game authorization required");
  return { ...state, players: [...state.players] as [string, string], board: [...state.board] };
}
