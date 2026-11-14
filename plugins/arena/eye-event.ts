import type { DurableMatchResult, EyeEvent } from "./types.js";
import { validateMatchResult } from "./validate-result.js";

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
