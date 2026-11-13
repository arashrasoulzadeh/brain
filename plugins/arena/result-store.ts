import type { DurableMatchResult } from "./types.js";
import { validateMatchResult } from "./validate-result.js";

export class MatchResultStore {
  private readonly results = new Map<string, DurableMatchResult>();
  save(result: DurableMatchResult): void {
    validateMatchResult(result);
    if (this.results.has(result.resultId)) throw new Error("result already exists");
    this.results.set(result.resultId, { ...result, winnerIds: [...result.winnerIds] });
  }
  get(resultId: string): DurableMatchResult | undefined {
    const result = this.results.get(resultId);
    return result ? { ...result, winnerIds: [...result.winnerIds] } : undefined;
  }
}
