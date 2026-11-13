import type { DurableMatchResult } from "./types.js";

export function validateMatchResult(result: DurableMatchResult): void {
  if (!/^[a-zA-Z0-9][a-zA-Z0-9_-]*:[a-zA-Z0-9][a-zA-Z0-9_-]*$/.test(result.resultId)) throw new Error("invalid result id");
  if (!result.projectId || !result.roomId || !result.completedAt || result.winnerIds.length === 0) throw new Error("invalid match result");
  if (!result.resultId.startsWith(`${result.projectId}:`)) throw new Error("result project mismatch");
}
