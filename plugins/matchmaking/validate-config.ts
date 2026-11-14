import type { MatchmakingConfig } from "./types.js";

export function validateConfig(config: MatchmakingConfig): void {
  if (
    !/^[a-zA-Z0-9][a-zA-Z0-9_-]*$/.test(config.projectId) ||
    !/^[a-zA-Z][a-zA-Z0-9_-]*$/.test(config.scope)
  )
    throw new Error("invalid matchmaking scope");
  if (
    config.matchSize < 1 ||
    config.matchSize > 64 ||
    config.ticketTtlSeconds < 1 ||
    config.ticketTtlSeconds > 3600
  )
    throw new Error("invalid matchmaking config");
}
