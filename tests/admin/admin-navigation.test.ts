import assert from "node:assert/strict";
import { test } from "node:test";
import { canViewAdminPage, enabledPluginIDs, visibleAdminPages } from "../../admin-nextjs/app/lib/admin-navigation.js";

test("admin navigation and route access use the same readiness policy", () => {
  const enabled = enabledPluginIDs([
    { PluginID: "matchmaking", Enabled: true, Ready: true },
    { PluginID: "tables", Enabled: true, Ready: false },
    { PluginID: "files", Enabled: false, Ready: true },
  ]);
  const pages = visibleAdminPages(3, enabled);
  assert.ok(pages.includes("live-sessions"));
  assert.ok(canViewAdminPage("live-sessions", enabled));
  assert.ok(pages.includes("monitoring"));
  assert.ok(!pages.includes("database"));
  assert.ok(!canViewAdminPage("database", enabled));
  assert.ok(!pages.includes("files"));
  assert.ok(!canViewAdminPage("files", enabled));
  assert.ok(!canViewAdminPage("workers", new Set()));
  assert.deepEqual(visibleAdminPages(0, new Set()), ["projects", "audit"]);
});
