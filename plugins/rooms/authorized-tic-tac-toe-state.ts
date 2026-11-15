import type { TicTacToeState } from "./index.js";
import { validateProjectId } from "../sdk-typescript/src/index.js";
export function authorizedTicTacToeState(
  state: TicTacToeState,
  projectId: string,
  playerId: string,
): TicTacToeState {
  validateProjectId(projectId, "game authorization required");
  if (state.projectId !== projectId || !state.players.includes(playerId))
    throw new Error("game authorization required");
  return { ...state, players: [...state.players] as [string, string], board: [...state.board] };
}
