import type { DurableMatchResult } from "./types.js";

export class MatchResultReader {
  constructor(private readonly results: ReadonlyMap<string, DurableMatchResult>) {}
  get(resultId: string): DurableMatchResult | undefined { const result = this.results.get(resultId); return result ? { ...result, winnerIds: [...result.winnerIds] } : undefined; }
}
