import type { DurableMatchResult } from "./types.js";
import { validateMatchResult } from "./validate-result.js";

export class MatchResultSaver {
  constructor(private readonly results: Map<string, DurableMatchResult>) {}
  save(result: DurableMatchResult): void {
    validateMatchResult(result);
    if (this.results.has(result.resultId)) throw new Error("result already exists");
    this.results.set(result.resultId, { ...result, winnerIds: [...result.winnerIds] });
  }
}
