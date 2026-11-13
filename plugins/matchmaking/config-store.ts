import type { MatchmakingConfig } from "./types.js";
import { validateConfig } from "./validate-config.js";

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
  isEnabled(projectId: string, scope: string): boolean { return this.configs.get(`${projectId}:${scope}`)?.enabled ?? false; }
}
