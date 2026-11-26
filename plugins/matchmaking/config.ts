import { type MatchmakingConfig } from "./types.js";
import { validateProjectId } from "../sdk-typescript/src/index.js";

export class MatchmakingConfigReader {
  constructor(private readonly configs: ReadonlyMap<string, MatchmakingConfig>) {}
  get(projectId: string, scope: string): MatchmakingConfig {
    validateProjectId(projectId);
    const config = this.configs.get(`${projectId}:${scope}`);
    if (!config) throw new Error("matchmaking config not found");
    return { ...config };
  }
}

export class MatchmakingConfigStatus {
  constructor(private readonly configs: ReadonlyMap<string, MatchmakingConfig>) {}
  isEnabled(projectId: string, scope: string): boolean {
    validateProjectId(projectId);
    return this.configs.get(`${projectId}:${scope}`)?.enabled ?? false;
  }
}

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

export class MatchmakingConfigWriter {
  constructor(private readonly configs: Map<string, MatchmakingConfig>) {}
  set(config: MatchmakingConfig): MatchmakingConfig {
    validateConfig(config);
    const key = `${config.projectId}:${config.scope}`;
    this.configs.set(key, { ...config });
    return { ...config };
  }
}
