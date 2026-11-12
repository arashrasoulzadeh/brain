import { request } from "../sdk-typescript/src/index.js";

export type MatchmakingConfig = { projectId: string; scope: string; matchSize: number; ticketTtlSeconds: number; enabled: boolean };

export function validateConfig(config: MatchmakingConfig): void {
  if (!/^[a-zA-Z0-9][a-zA-Z0-9_-]*$/.test(config.projectId) || !/^[a-zA-Z][a-zA-Z0-9_-]*$/.test(config.scope)) throw new Error("invalid matchmaking scope");
  if (config.matchSize < 1 || config.matchSize > 64 || config.ticketTtlSeconds < 1 || config.ticketTtlSeconds > 3600) throw new Error("invalid matchmaking config");
}

export class MatchmakingConfigs {
  private readonly configs = new Map<string, MatchmakingConfig>();
  set(config: MatchmakingConfig): MatchmakingConfig {
    validateConfig(config);
    const key = `${config.projectId}:${config.scope}`;
    this.configs.set(key, { ...config });
    return { ...config };
  }
  get(projectId: string, scope: string): MatchmakingConfig {
    const config = this.configs.get(`${projectId}:${scope}`);
    if (!config) throw new Error("matchmaking config not found");
    return { ...config };
  }
}

export function matchmakingRequest(projectId: string, pluginId: string, version: string, operation: "matchmaking.queue" | "matchmaking.cancel" | "matchmaking.status" | "matchmaking.assign", config: string, payload: unknown = {}) {
  return request(`matchmaking-${operation}-${projectId}-${config}`, { protocol: 1, projectId, pluginId, version, operation, payload: { config, ...(payload as Record<string, unknown>) } });
}

