import type { ArenaRuntime } from "./arena-state.js";
import type { ArenaState } from "./types.js";
export class ArenaStateReader {
  constructor(private readonly runtime: ArenaRuntime) {}
  read(): ArenaState {
    return this.runtime.state;
  }
}
