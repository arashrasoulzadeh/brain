import type { TicTacToeRuntime } from "./tic-tac-toe-runtime.js";
export class TicTacToeRevisionReader { constructor(private readonly runtime: TicTacToeRuntime) {} read(): number { return this.runtime.revision; } }
