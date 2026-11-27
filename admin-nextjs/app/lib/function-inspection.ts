export type FunctionLog = {
  status: string;
  duration_ms: number;
  created_at: string;
  message?: string;
  reason?: string;
  error?: string;
  method?: string;
  request_id?: string;
  steps?: { name: string; status: string; duration_ms: number; reason?: string; error?: string }[];
};
export type LogOutcome = "all" | "success" | "failure";
export function filterFunctionLogs(logs: readonly FunctionLog[], query: string, outcome: LogOutcome): FunctionLog[] {
  const needle = query.trim().toLowerCase();
  return logs.slice().reverse().filter(log =>
    (outcome === "all" || (log.status.toLowerCase() === "success") === (outcome === "success")) &&
    [log.status, log.created_at, log.method, log.request_id, log.reason, log.message, log.error].some(value => value?.toLowerCase().includes(needle)),
  ).sort((a, b) => (Date.parse(b.created_at) || 0) - (Date.parse(a.created_at) || 0));
}
export function functionLogMetrics(logs: readonly FunctionLog[]) {
  const durations = logs.map(log => log.duration_ms).filter(value => Number.isFinite(value) && value >= 0).sort((a, b) => a - b);
  const successful = logs.filter(log => log.status.toLowerCase() === "success").length;
  return { successful, average: durations.length ? Math.round(durations.reduce((sum, value) => sum + value, 0) / durations.length) : null,
    p95: durations.length ? durations[Math.ceil(durations.length * 0.95) - 1] : null };
}
export function filterFunctionDrafts<T extends { name: string; public?: boolean; published_at?: string | null }>(drafts: readonly T[], query: string, visibility: string, publication: string): T[] {
  const needle = query.trim().toLowerCase();
  return drafts.filter(draft => draft.name.toLowerCase().includes(needle) &&
    (!visibility || Boolean(draft.public) === (visibility === "public")) &&
    (!publication || Boolean(draft.published_at) === (publication === "published")))
    .sort((a, b) => a.name.localeCompare(b.name));
}
