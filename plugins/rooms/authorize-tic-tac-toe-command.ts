import { authorizedTicTacToeState } from "./authorized-tic-tac-toe-state.js";
import type { TicTacToeState } from "./index.js";
export function authorizeTicTacToeCommand(
  state: TicTacToeState,
  projectId: string,
  playerId: string,
  operation: string,
): boolean {
  authorizedTicTacToeState(state, projectId, playerId);
  if (operation !== "move") throw new Error("game command not authorized");
  return true;
}
