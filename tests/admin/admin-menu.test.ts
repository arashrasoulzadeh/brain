import assert from "node:assert/strict";
import { test } from "node:test";
import { adminMenu, menuItemActive } from "../../admin-nextjs/app/lib/admin-menu.js";
import type { AdminContribution } from "../../admin-nextjs/app/lib/admin-registry.js";

const page: AdminContribution = { projectID: "one", pluginID: "custom", version: "1", href: "/extensions/one/custom/1/list", page: { id: "list", title: "Records", actions: [] } };
test("compact menu preserves gated routes and discovers extensions and aliases", () => {
  const menu = adminMenu(["projects", "audit"], [page, { ...page, href: "/other", page: { ...page.page, legacy_route: "live-sessions" } }]);
  assert.deepEqual(menu.flatMap(group => group.items.map(item => item.href)), ["/projects", "/live-sessions", page.href, "/audit"]);
  assert.deepEqual(adminMenu(["projects", "audit"], [page], "custom@1").flatMap(group => group.items.map(item => item.href)), [page.href]);
  assert.deepEqual(adminMenu([], [], "missing"), []);
});
test("navigation matches nested built-in routes and exact extension pages", () => {
  assert.equal(menuItemActive("/", "/projects"), true);
  assert.equal(menuItemActive("/functions/new", "/functions"), true);
  assert.equal(menuItemActive("/functions2", "/functions"), false);
  assert.equal(menuItemActive(page.href + "/other", page.href), false);
  assert.equal(menuItemActive(page.href, page.href), true);
});
