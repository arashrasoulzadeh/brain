import { validateTicTacToeMove } from "./validate-tic-tac-toe-move.js";
import type { TicTacToeRuntime } from "./tic-tac-toe-runtime.js";
export class TicTacToeMover {
  constructor(private readonly runtime: TicTacToeRuntime) {}
  move(playerId: string, cell: number, expectedRevision: number) {
    if (expectedRevision !== this.runtime.revision) throw new Error("stale game revision");
    this.runtime.current = validateTicTacToeMove(this.runtime.current, playerId, cell);
    this.runtime.revision += 1;
    return {
      ...this.runtime.current,
      players: [...this.runtime.current.players] as [string, string],
      board: [...this.runtime.current.board],
    };
  }
}
