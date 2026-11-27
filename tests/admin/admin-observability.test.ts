import assert from "node:assert/strict";
import { test } from "node:test";
import { filterAudit, filterInventory, installationHealth } from "../../admin-nextjs/app/lib/admin-observability.js";

test("audit filters by outcome and source, orders actual times, preserves input", () => {
  const records = [
    { action: "grant", source: "operator", success: true, occurred_at: "2026-01-02T00:00:00Z", version: "2" },
    { action: "deny", source: "operator", success: false, occurred_at: "2026-01-03T00:00:00Z" },
    { action: "install", source: "plugin", success: true, occurred_at: "2026-01-01T00:00:00Z", detail: "demo" },
  ];
  assert.deepEqual(filterAudit(records, "", "all", "").map(x => x.action), ["deny", "grant", "install"]);
  assert.deepEqual(filterAudit(records, "", "failure", "operator").map(x => x.action), ["deny"]);
  assert.deepEqual(filterAudit(records, " DEMO ", "success", "plugin").map(x => x.action), ["install"]);
  assert.deepEqual(filterAudit(records, "2", "success", "operator").map(x => x.action), ["grant"]);
  assert.deepEqual(filterAudit(records, "missing", "all", ""), []);
  assert.equal(records[0].action, "grant");
});
test("inventory filtering matches resource keys and file metadata", () => {
  const rows = [{ key: "a", filename: "Photo.PNG", content_type: "image/png" }, { key: "b" }];
  assert.deepEqual(filterInventory(rows, " PNG "), [rows[0]]);
  assert.deepEqual(filterInventory(rows, "b"), [rows[1]]);
  assert.deepEqual(filterInventory(rows, ""), rows);
  assert.deepEqual(filterInventory(rows, "missing"), []);
});
test("worker health never treats missing readiness or grants as healthy", () => {
  const manifest = { ID: "files", Title: "Files", Description: "", Version: "1", Capabilities: ["read"] };
  const status = { ProjectID: "one", PluginID: "files", Version: "1", Enabled: true, Ready: true, WorkerReady: true, Draining: false, Grants: ["read"] };
  assert.equal(installationHealth(manifest), "absent");
  assert.equal(installationHealth(manifest, { ...status, Draining: true }), "draining");
  assert.equal(installationHealth(manifest, { ...status, Ready: false }), "pending");
  assert.equal(installationHealth(manifest, { ...status, Enabled: false }), "disabled");
  assert.equal(installationHealth(manifest, { ...status, Grants: [] }), "grants");
  assert.equal(installationHealth(manifest, { ...status, WorkerReady: undefined }), "worker");
  assert.equal(installationHealth(manifest, { ...status, WorkerReady: false }), "worker");
  assert.equal(installationHealth(manifest, status), "ready");
});
