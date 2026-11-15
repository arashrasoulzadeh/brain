import type { MatchmakingConfig } from "./types.js";
import { validateProjectId } from "../sdk-typescript/src/index.js";

export class MatchmakingConfigStatus {
  constructor(private readonly configs: ReadonlyMap<string, MatchmakingConfig>) {}
  isEnabled(projectId: string, scope: string): boolean {
    validateProjectId(projectId);
    return this.configs.get(`${projectId}:${scope}`)?.enabled ?? false;
  }
}
