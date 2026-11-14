export type DurableMatchResult = {
  resultId: string;
  projectId: string;
  roomId: string;
  winnerIds: readonly string[];
  completedAt: string;
};
export type EyeEvent = {
  projectId: string;
  name: "arena.result";
  fields: { resultId: string; roomId: string; winnerCount: number };
};
export type ArenaState = "waiting" | "active" | "finished" | "interrupted";
export type ArenaMeasurement = { iterations: number; elapsedMs: number; memoryBytes: number };
