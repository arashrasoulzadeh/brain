import assert from "node:assert/strict";
import { test } from "node:test";
import { adminMenu, menuItemActive } from "../../admin-nextjs/app/lib/admin-menu.js";
import { adminWorkspaceFallbacks, type AdminContribution } from "../../admin-nextjs/app/lib/admin-registry.js";

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

test("each enabled operational plugin has its own project sidebar destination", () => {
  const aliases = ["rooms", "matchmaking", "arena", "invitations", "subscriptions", "auth", "access-management", "analytics", "games", "pvp", "database", "files"];
  const extensions = aliases.map(alias => ({ ...page, pluginID: alias, href: "/one/ui/extensions/" + alias + "/1/operations", page: { ...page.page, legacy_route: alias } }));
  const items = adminMenu(["projects", "functions", "plugins"], extensions, "", "one").flatMap(group => group.items);
  for (const alias of aliases) assert.equal(items.filter(item => item.href === "/one/ui/" + alias).length, 1);
  assert.equal(items.find(item => item.href === "/one/ui/pvp")?.label, "PvP");
  assert.equal(items.find(item => item.href === "/one/ui/auth")?.label, "Authentication");
  assert.equal(items.some(item => item.href === "/two/ui/files"), false);
  assert.equal(items.some(item => item.href === "/one/ui/live-sessions"), false);
});

test("enabled plugins without manifest action pages have a project-scoped workspace and disappear when disabled", () => {
  const manifests = ["dashboard", "custom", "functions"].map(ID => ({ ID, Title: ID, Description: "", Version: "1", Capabilities: ["read"], Routes: [], Dependencies: [], ExtensionPoints: [], SourceFiles: [], Imports: [], PluginImports: [] }));
  const installations = manifests.map(m => ({ ProjectID: "one", PluginID: m.ID, Version: "1", Enabled: true, Ready: true, Draining: false, Grants: ["read"] }));
  const catalog = { manifests, installations };
  const links = adminWorkspaceFallbacks("one", catalog, []);
  assert.deepEqual(links.map(item => item.href), ["/one/ui/dashboard", "/one/ui/plugins/custom%401"]);
  assert.deepEqual(adminWorkspaceFallbacks("two", catalog, []), []);
  installations[0].Enabled = false;
  assert.deepEqual(adminWorkspaceFallbacks("one", catalog, []).map(item => item.href), ["/one/ui/plugins/custom%401"]);
});
