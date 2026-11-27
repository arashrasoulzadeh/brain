import { requireProjectInput, type PluginContext, type TurnMatchState, type CommittedCommand } from "../sdk-typescript/src/index.js";

export type TicTacToeState = TurnMatchState & {
  players: readonly [string, string];
  board: readonly ("X" | "O" | null)[];
};
export type CommittedTicTacToeMove = CommittedCommand & { cell: number };
export { initializeTicTacToe } from "./tic-tac-toe.js";
export { validateTicTacToeMove } from "./tic-tac-toe.js";
export { expireTicTacToeTurn } from "./recovery.js";
export { ticTacToeWinner } from "./tic-tac-toe.js";
export { ticTacToeDraw } from "./tic-tac-toe.js";
export { validateTicTacToeRuleResult } from "./tic-tac-toe.js";
export { pauseTicTacToeTurn } from "./recovery.js";
export { resumePausedTicTacToeTurn } from "./recovery.js";
export { reconstructTicTacToeHistory } from "./recovery.js";
export { TIC_TAC_TOE_RULES_SOURCE, publishTicTacToeRules } from "./tic-tac-toe.js";
export { authorizedTicTacToeState } from "./recovery.js";
export { authorizeTicTacToeCommand } from "./recovery.js";
export { validateTicTacToeCommandPayload } from "./recovery.js";
export { createTicTacToeCoordinator } from "./tic-tac-toe.js";
export { TicTacToeRevisionReader } from "./tic-tac-toe.js";
export { TicTacToeMover } from "./tic-tac-toe.js";
export { serializeTicTacToeState } from "./recovery.js";
export { reconnectTicTacToe } from "./recovery.js";
export { hiddenTicTacToeView } from "./recovery.js";

export function modes(context: PluginContext, input: unknown) {
  requireProjectInput(context, input);
  return { projectId: context.project.id, modes: ["tic-tac-toe"] };
}
