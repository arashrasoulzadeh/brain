import { request } from "../sdk-typescript/src/index.js";

export type DurableMatchResult = {
  resultId: string;
  projectId: string;
  roomId: string;
  winnerIds: readonly string[];
  completedAt: string;
};

export function validateMatchResult(result: DurableMatchResult): void {
  if (!/^[a-zA-Z0-9][a-zA-Z0-9_-]*:[a-zA-Z0-9][a-zA-Z0-9_-]*$/.test(result.resultId)) throw new Error("invalid result id");
  if (!result.projectId || !result.roomId || !result.completedAt || result.winnerIds.length === 0) throw new Error("invalid match result");
  if (!result.resultId.startsWith(`${result.projectId}:`)) throw new Error("result project mismatch");
}

export function arenaRequest(projectId: string, pluginId: string, version: string, operation: "arena.start" | "arena.input" | "arena.snapshot" | "arena.result", roomId: string, payload: unknown = {}) {
  return request(`arena-${operation}-${projectId}-${roomId}`, { protocol: 1, projectId, pluginId, version, operation, payload: { roomId, ...(payload as Record<string, unknown>) } });
}

