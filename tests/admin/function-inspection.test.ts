import assert from "node:assert/strict";
import { test } from "node:test";
import { filterFunctionDrafts, filterFunctionLogs, functionLogMetrics } from "../../admin-nextjs/app/lib/function-inspection.js";
test("function list filters name, publication and visibility without changing input", () => {
  const drafts = [{ name: "Zulu" }, { name: "Alpha", public: true, published_at: "2026-01-01" }, { name: "Beta", public: true }];
  assert.deepEqual(filterFunctionDrafts(drafts, "", "", "").map(x => x.name), ["Alpha", "Beta", "Zulu"]);
  assert.deepEqual(filterFunctionDrafts(drafts, " AL ", "public", "published"), [drafts[1]]);
  assert.deepEqual(filterFunctionDrafts(drafts, "", "private", "draft"), [drafts[0]]);
  assert.deepEqual(filterFunctionDrafts(drafts, "missing", "", ""), []);
  assert.equal(drafts[0].name, "Zulu");
});
test("log filtering and metrics use the loaded execution records", () => {
  const logs = [
    { status: "SUCCESS", duration_ms: 10, created_at: "2026-01-01", request_id: "abc" },
    { status: "failure", duration_ms: 90, created_at: "2026-01-03", error: "Denied" },
    { status: "success", duration_ms: 20, created_at: "2026-01-02", method: "GET" },
  ];
  assert.deepEqual(filterFunctionLogs(logs, "", "all"), [logs[1], logs[2], logs[0]]);
  assert.deepEqual(filterFunctionLogs(logs, " DENIED ", "failure"), [logs[1]]);
  assert.deepEqual(filterFunctionLogs(logs, "abc", "success"), [logs[0]]);
  assert.deepEqual(filterFunctionLogs(logs, "GET", "success"), [logs[2]]);
  assert.deepEqual(functionLogMetrics(logs), { successful: 2, average: 40, p95: 90 });
  assert.deepEqual(functionLogMetrics([]), { successful: 0, average: null, p95: null });
});
