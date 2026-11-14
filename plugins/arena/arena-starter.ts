import type { ArenaRuntime } from "./arena-state.js";
export class ArenaStarter {
  constructor(private readonly runtime: ArenaRuntime) {}
  start(players: readonly string[]): void {
    if (this.runtime.state !== "waiting") throw new Error("arena already started");
    if (
      players.length < 2 ||
      new Set(players).size !== players.length ||
      players.some((player) => !player.trim())
    )
      throw new Error("invalid arena players");
    this.runtime.players = [...players];
    this.runtime.state = "active";
  }
}
