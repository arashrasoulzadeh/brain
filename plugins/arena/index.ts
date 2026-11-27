import { createResourceRequest } from "../sdk-typescript/src/index.js";

export const arenaRequest = createResourceRequest<"arena.start" | "arena.input" | "arena.snapshot" | "arena.result" | "arena.connect" | "arena.player.input">("arena", "roomId");

export { arenaStartCommand, start, input, snapshot, result, connect, playerInput, interrupt } from "./state-machine.js";
export { exportEyeEvent } from "./match-results.js";
export { MatchResultSaver } from "./match-results.js";
export { MatchResultReader } from "./match-results.js";
export { createArena } from "./state-machine.js";
export { ArenaStateReader } from "./state-machine.js";
export { ArenaStarter } from "./state-machine.js";
export { ArenaActionRecorder } from "./state-machine.js";
export { ArenaMeasurer } from "./state-machine.js";
export { ArenaFinisher } from "./state-machine.js";
export { interruptExpiredActionRoom } from "./state-machine.js";
export { validateMatchResult } from "./match-results.js";
export type { ArenaMeasurement, ArenaState, DurableMatchResult, EyeEvent } from "./types.js";
