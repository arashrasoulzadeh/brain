import type { TicTacToeState } from "./index.js";
export function serializeTicTacToeState(state: TicTacToeState): string { return JSON.stringify(state); }
