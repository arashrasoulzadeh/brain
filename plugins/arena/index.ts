import { pluginRequest } from "../sdk-typescript/src/index.js";

export function arenaRequest(
  projectId: string,
  pluginId: string,
  version: string,
  operation: "arena.start" | "arena.input" | "arena.snapshot" | "arena.result",
  roomId: string,
  payload: unknown = {},
) {
  return pluginRequest(
    "arena",
    projectId,
    pluginId,
    version,
    operation,
    { roomId, ...(payload as Record<string, unknown>) },
    roomId,
  );
}

export { arenaStartCommand } from "./state-machine.js";
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
