import type { MatchmakingTicket } from "./types.js";
import type { MatchmakingConfigReader } from "./matchmaking-config-reader.js";
import type { MatchmakingConfigStatus } from "./matchmaking-config-status.js";
export type QueueRuntime = {
  sequence: number;
  decisionSequence: number;
  lifecycleRevision: number;
  pausedProjects: Set<string>;
  queues: Map<string, MatchmakingTicket[]>;
  configs?: { reader: MatchmakingConfigReader; status: MatchmakingConfigStatus };
};
export function createQueueRuntime(configs?: QueueRuntime["configs"]): QueueRuntime {
  return {
    sequence: 0,
    decisionSequence: 0,
    lifecycleRevision: 0,
    pausedProjects: new Set(),
    queues: new Map(),
    configs,
  };
}
