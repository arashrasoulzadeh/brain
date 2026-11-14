import { createTicTacToeRuntime } from "./tic-tac-toe-runtime.js";
import { TicTacToeRevisionReader } from "./tic-tac-toe-revision-reader.js";
import { TicTacToeMover } from "./tic-tac-toe-mover.js";
import type { TicTacToeState } from "./index.js";
export function createTicTacToeCoordinator(state: TicTacToeState) { const runtime = createTicTacToeRuntime(state); const revision = new TicTacToeRevisionReader(runtime); const mover = new TicTacToeMover(runtime); return { revision: revision.read.bind(revision), move: mover.move.bind(mover) }; }
