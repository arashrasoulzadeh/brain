import assert from "node:assert/strict";
import { test } from "node:test";
import { functionTemplate, functionMethods, functionNameError, functionVersion, functionPermissions, functionInput } from "../../admin-nextjs/app/lib/function-studio.js";
import { filterFunctionDrafts, filterFunctionLogs, functionLogMetrics, projectFunctionRecords, functionCatalogSummary, functionCatalogRows, functionCatalogPage } from "../../admin-nextjs/app/lib/function-inspection.js";
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

test("function catalog counts only authorized project records and keeps publication independent of visibility", () => {
  const records = [
    { project_id: "dev", name: "alpha", source: "secret source", created_at: "2026-01-01", public: true, method: "GET", version: 1 },
    { project_id: "dev", name: "beta", created_at: "2026-01-02", published_at: "2026-01-03", public: false, method: "POST", version: 3 },
    { project_id: "other", name: "outside", created_at: "2026-01-04", public: true },
  ];
  const scoped = projectFunctionRecords(records, "dev");
  assert.equal(scoped.length, 2);
  assert.equal("source" in scoped[0], false);
  assert.deepEqual(functionCatalogSummary(scoped), { total: 2, published: 1, unpublished: 1, public: 1 });
  assert.deepEqual(functionCatalogRows(scoped, { query: "", visibility: "private", publication: "published", method: "POST", sort: "newest" }).map(row => row.name), ["beta"]);
  assert.deepEqual(functionCatalogRows(scoped, { query: "", visibility: "", publication: "", method: "", sort: "version" }).map(row => row.name), ["beta", "alpha"]);
  assert.deepEqual(functionCatalogPage(scoped, 9, 1), { page: 1, pages: 2, start: 2, end: 2, rows: [scoped[1]] });
  assert.deepEqual(functionCatalogPage([], 9, 10), { page: 0, pages: 1, start: 0, end: 0, rows: [] });
  assert.equal(records[0].name, "alpha");
});

test("studio templates and input checks follow the host function contract", () => {
  for (const method of functionMethods) {
    for (const template of ["echo", "health", "transform"] as const) {
      const source = functionTemplate(template, method);
      assert.match(source, new RegExp("method: '" + method + "'"));
      assert.match(source, /export default async function handler/);
    }
  }
  assert.equal(functionNameError("hello-world"), "");
  assert.notEqual(functionNameError("2bad"), "");
  assert.notEqual(functionNameError("bad/name"), "");
  assert.equal(functionVersion("3"), 3);
  assert.equal(functionVersion("1.5"), undefined);
  assert.equal(functionVersion("0"), undefined);
  assert.deepEqual(functionPermissions("functions.invoke, , functions.invoke, tables.read"), ["functions.invoke", "tables.read"]);
  assert.deepEqual(functionInput('{"x":1}'), { x: 1 });
  assert.equal(functionInput("null"), null);
  assert.throws(() => functionInput("{"), /valid JSON/);
});
