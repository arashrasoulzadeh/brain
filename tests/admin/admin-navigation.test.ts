import assert from "node:assert/strict";
import { test } from "node:test";
import { canViewAdminPage, enabledPluginIDs, visibleAdminPages } from "../../admin-nextjs/app/lib/admin-navigation.js";

test("platform and specialized navigation retain readiness policy", () => {
  const enabled = enabledPluginIDs([
    { PluginID: "matchmaking", Enabled: true, Ready: true },
    { PluginID: "functions", Enabled: true, Ready: false },
  ]);
  const pages = visibleAdminPages(2, enabled);
  assert.ok(pages.includes("monitoring"));
  assert.ok(!pages.includes("functions"));
  assert.ok(!canViewAdminPage("functions", enabled));
  assert.ok(!pages.includes("live-sessions"), "plugin aliases are now discovered from manifests");
  assert.ok(canViewAdminPage("workers", new Set()));
  assert.deepEqual(visibleAdminPages(0, new Set()), ["projects", "audit", "workers"]);
});
