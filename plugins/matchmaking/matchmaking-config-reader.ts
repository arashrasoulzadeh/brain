import type { MatchmakingConfig } from "./types.js";
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
