import assert from "node:assert/strict";
import { test } from "node:test";
import { adminMenu, menuItemActive } from "../../admin-nextjs/app/lib/admin-menu.js";
import type { AdminContribution } from "../../admin-nextjs/app/lib/admin-registry.js";

const page: AdminContribution = { projectID: "one", pluginID: "custom", version: "1", href: "/one/ui/extensions/custom/1/list", page: { id: "list", title: "Records", actions: [] } };
test("compact menu preserves gated routes and discovers extensions and aliases", () => {
  const menu = adminMenu(["projects", "audit"], [page, { ...page, href: "/other", page: { ...page.page, legacy_route: "live-sessions" } }], "", "one");
  assert.deepEqual(menu.flatMap(group => group.items.map(item => item.href)), ["/_system/ui/projects", "/one/ui/live-sessions", page.href, "/_system/ui/audit"]);
  assert.deepEqual(adminMenu(["projects", "audit"], [page], "custom@1", "one").flatMap(group => group.items.map(item => item.href)), [page.href]);
  assert.deepEqual(adminMenu([], [], "missing"), []);
});
test("navigation matches nested built-in routes and exact extension pages", () => {
  assert.equal(menuItemActive("/_system/ui", "/_system/ui/projects"), true);
  assert.equal(menuItemActive("/one/ui/functions/new", "/one/ui/functions"), true);
  assert.equal(menuItemActive("/one/ui/functions2", "/one/ui/functions"), false);
  assert.equal(menuItemActive(page.href + "/other", page.href), false);
  assert.equal(menuItemActive(page.href, page.href), true);
});

test("project links are not generated without a project and never match another project", () => {
  const links = adminMenu(["projects", "plugins", "functions", "audit"], [page], "", "two").flatMap(group => group.items.map(item => item.href));
  assert.deepEqual(links, ["/_system/ui/projects", "/two/ui/plugins", "/two/ui/functions", "/_system/ui/audit"]);
  assert.deepEqual(adminMenu(["projects", "plugins", "functions", "audit"], [page]).flatMap(group => group.items.map(item => item.href)), ["/_system/ui/projects", "/_system/ui/audit"]);
  assert.equal(menuItemActive("/two/ui/functions/new", "/one/ui/functions"), false);
});
