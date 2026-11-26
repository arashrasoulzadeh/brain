import { validateProjectId } from "../sdk-typescript/src/index.js";
import type { DurableMatchResult, EyeEvent } from "./types.js";

export function validateMatchResult(result: DurableMatchResult): void {
  validateProjectId(result.projectId, "invalid match result");
  validateProjectId(result.roomId, "invalid match result");
  if (!/^[a-zA-Z0-9][a-zA-Z0-9_-]*:[a-zA-Z0-9][a-zA-Z0-9_-]*$/.test(result.resultId))
    throw new Error("invalid result id");
  if (
    !result.completedAt ||
    !Number.isFinite(Date.parse(result.completedAt)) ||
    result.winnerIds.length === 0
  )
    throw new Error("invalid match result");
  if (!result.resultId.startsWith(`${result.projectId}:`))
    throw new Error("result project mismatch");
}

export class MatchResultSaver {
  constructor(private readonly results: Map<string, DurableMatchResult>) {}
  save(result: DurableMatchResult): void {
    validateMatchResult(result);
    if (this.results.has(result.resultId)) throw new Error("result already exists");
    this.results.set(result.resultId, { ...result, winnerIds: [...result.winnerIds] });
  }
}

export class MatchResultReader {
  constructor(private readonly results: ReadonlyMap<string, DurableMatchResult>) {}
  get(resultId: string): DurableMatchResult | undefined {
    const result = this.results.get(resultId);
    return result ? { ...result, winnerIds: [...result.winnerIds] } : undefined;
  }
}

export function exportEyeEvent(result: DurableMatchResult): EyeEvent {
  validateMatchResult(result);
  return {
    projectId: result.projectId,
    name: "arena.result",
    fields: {
      resultId: result.resultId,
      roomId: result.roomId,
      winnerCount: result.winnerIds.length,
    },
  };
}
