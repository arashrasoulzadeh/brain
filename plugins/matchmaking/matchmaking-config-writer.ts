import type { MatchmakingConfig } from "./types.js";
import { validateConfig } from "./validate-config.js";

export class MatchmakingConfigWriter {
  constructor(private readonly configs: Map<string, MatchmakingConfig>) {}
  set(config: MatchmakingConfig): MatchmakingConfig {
    validateConfig(config);
    const key = `${config.projectId}:${config.scope}`;
    this.configs.set(key, { ...config });
    return { ...config };
  }
}
