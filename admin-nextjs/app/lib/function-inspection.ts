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
export function filterFunctionDrafts<T extends { name: string; public?: boolean; published_at?: string }>(drafts: readonly T[], query: string, visibility: string, publication: string): T[] {
  const needle = query.trim().toLowerCase();
  return drafts.filter(draft => draft.name.toLowerCase().includes(needle) &&
    (!visibility || Boolean(draft.public) === (visibility === "public")) &&
    (!publication || Boolean(draft.published_at) === (publication === "published")))
    .sort((a, b) => a.name.localeCompare(b.name));
}

export type FunctionCatalogRecord = {
  project_id: string;
  name: string;
  version?: number;
  public?: boolean;
  permissions?: string[] | null;
  method?: string;
  created_at: string;
  published_at?: string;
};
export type FunctionCatalogSort = "name" | "newest" | "published" | "version";
export type FunctionCatalogFilters = { query: string; visibility: string; publication: string; method: string; sort: FunctionCatalogSort };

export function projectFunctionRecords(records: readonly FunctionCatalogRecord[], projectID: string): FunctionCatalogRecord[] {
  return records.filter(record => record.project_id === projectID).map(record => ({
    project_id: record.project_id, name: record.name, version: record.version, public: record.public,
    permissions: record.permissions, method: record.method, created_at: record.created_at, published_at: record.published_at,
  }));
}

export function functionCatalogSummary(records: readonly FunctionCatalogRecord[]) {
  const published = records.filter(record => Boolean(record.published_at)).length;
  return { total: records.length, published, unpublished: records.length - published, public: records.filter(record => record.public).length };
}

export function functionCatalogRows(records: readonly FunctionCatalogRecord[], filters: FunctionCatalogFilters): FunctionCatalogRecord[] {
  const rows = filterFunctionDrafts(records, filters.query, filters.visibility, filters.publication)
    .filter(record => !filters.method || record.method === filters.method);
  const timestamp = (value?: string) => Date.parse(value ?? "") || 0;
  return rows.sort((a, b) => {
    const order = filters.sort === "newest" ? timestamp(b.created_at) - timestamp(a.created_at)
      : filters.sort === "published" ? timestamp(b.published_at) - timestamp(a.published_at)
      : filters.sort === "version" ? (b.version ?? 1) - (a.version ?? 1) : 0;
    return order || a.name.localeCompare(b.name);
  });
}

export function functionCatalogPage<T>(records: readonly T[], requestedPage: number, pageSize: number) {
  const pages = Math.max(1, Math.ceil(records.length / pageSize));
  const page = Math.max(0, Math.min(requestedPage, pages - 1));
  const start = page * pageSize;
  return { page, pages, start: records.length ? start + 1 : 0, end: Math.min(start + pageSize, records.length), rows: records.slice(start, start + pageSize) };
}
