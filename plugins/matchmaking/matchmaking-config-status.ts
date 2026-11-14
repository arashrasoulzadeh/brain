import type { MatchmakingConfig } from "./types.js";

export class MatchmakingConfigStatus {
  constructor(private readonly configs: ReadonlyMap<string, MatchmakingConfig>) {}
  isEnabled(projectId: string, scope: string): boolean {
    return this.configs.get(`${projectId}:${scope}`)?.enabled ?? false;
  }
}
