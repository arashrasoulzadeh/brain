import type { DurableMatchResult } from "./types.js";
import { validateProjectId } from "../sdk-typescript/src/index.js";

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
