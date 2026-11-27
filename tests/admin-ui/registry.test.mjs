import assert from "node:assert/strict";
import { createRequire } from "node:module";
import { test } from "node:test";

const require = createRequire(new URL("../../admin-nextjs/package.json", import.meta.url));
const { enabledPluginIDs, canViewAdminPage, visibleAdminPages } = require("./app/lib/admin-navigation.ts");
const { adminContributions, adminPageExtensions, adminLegacyContributions, adminEmbeddedExtensions, adminLegacyRootPages } = require("./app/lib/admin-registry.ts");
const page = (id, title, extra = {}) => ({ id, title, actions: [{ operation: "read", title: "Read", read_only: true }, { operation: "missing", title: "Missing" }], ...extra });
const manifest = (ID, pages, extra = {}) => ({ ID, Version: "1", Capabilities: ["read"], Routes: [{ Operation: "read", Capability: "read" }], Admin: { version: 1, pages }, ...extra });
const installation = (PluginID, extra = {}) => ({ PluginID, ProjectID: "dev", Version: "1", Enabled: true, Ready: true, Grants: ["read"], ...extra });

test("navigation only presents ready enabled plugin-backed screens", () => {
  const enabled = enabledPluginIDs([installation("dashboard"), installation("functions"), installation("missing", { Ready: false }), installation("disabled", { Enabled: false })]);
  assert.deepEqual([...enabled], ["dashboard", "functions"]);
  assert.deepEqual(visibleAdminPages(1, enabled), ["projects", "plugins", "audit", "system", "functions", "routes", "monitoring", "workers"]);
  assert.deepEqual(visibleAdminPages(0, new Set()), ["projects", "audit"]);
  for (const page of ["system", "functions", "routes", "monitoring", "workers"]) assert.equal(canViewAdminPage(page, new Set()), false);
  assert.equal(canViewAdminPage("projects", new Set()), true);
});

test("contribution discovery checks exact project/version, lifecycle, all grants and route declarations", () => {
  const manifests = [
    manifest("sample", [page("z", "Z"), page("b", "A"), page("a", "A"), page("empty", "Empty", { actions: [{ operation: "undeclared" }] })]),
    manifest("foreign", [page("one", "Foreign")]),
    manifest("disabled", [page("one", "Disabled")]),
    manifest("pending", [page("one", "Pending")]),
    manifest("draining", [page("one", "Draining")]),
    manifest("partial", [page("one", "Partial")], { Capabilities: ["read", "write"] }),
    manifest("unsupported", [], { Admin: { version: 2, pages: [] } }),
    manifest("no-routes", [page("one", "No routes")], { Routes: undefined }),
    manifest("no-admin", [], { Admin: undefined }),
    manifest("wrong-version", [page("one", "Wrong version")]),
  ];
  const installations = [
    installation("sample"), installation("foreign", { ProjectID: "other" }), installation("disabled", { Enabled: false }),
    installation("pending", { Ready: false }), installation("draining", { Draining: true }), installation("partial"),
    installation("unsupported"), installation("no-routes"), installation("no-admin"), installation("wrong-version", { Version: "2" }),
    installation("sample", { Grants: undefined }), installation("absent"),
  ];
  const contributions = adminContributions("dev", { manifests, installations });
  assert.deepEqual(contributions.map(item => item.page.id), ["a", "b", "z"]);
  assert.deepEqual(contributions[0].page.actions.map(item => item.operation), ["read"]);
  assert.equal(contributions[0].href, "/extensions/dev/sample/1/a");
  assert.deepEqual(adminContributions("", { manifests, installations }), []);
  assert.deepEqual(adminContributions("dev", { manifests: null, installations }), []);
  assert.deepEqual(adminContributions("dev", { manifests, installations: null }), []);
  assert.equal(adminContributions("dev space", { manifests, installations: [installation("sample", { ProjectID: "dev space" })] })[0].href, "/extensions/dev%20space/sample/1/a");
});

test("extensions require matching project, explicit tab slots and target identity", () => {
  const contribution = (pluginID, page, projectID = "dev") => ({ projectID, pluginID, version: "1", page, href: "/" + pluginID + "/" + page.id });
  const target = contribution("base", page("main", "Base", { extension_slots: ["tab"], legacy_route: "files" }));
  const tab = contribution("tab", page("view", "Tab", { extends: "base/main", extends_slot: "tab", legacy_route: "files" }));
  const link = contribution("link", page("view", "Link", { extends: "base/main", legacy_route: "files" }));
  const other = contribution("foreign", page("view", "Other", { extends: "base/main", extends_slot: "tab" }), "other");
  const self = contribution("base", page("self", "Self", { extends: "base/main", extends_slot: "tab" }));
  const unrelated = contribution("unrelated", page("view", "Unrelated", { extends: "base/unknown" }));
  const contributions = [target, tab, link, other, self, unrelated];
  assert.deepEqual(adminPageExtensions(target, contributions), [tab, link]);
  assert.deepEqual(adminEmbeddedExtensions(target, contributions), [tab]);
  assert.deepEqual(adminEmbeddedExtensions({ ...target, page: { ...target.page, extension_slots: [] } }, contributions), []);
  assert.deepEqual(adminLegacyRootPages([target, tab, link]), [target, link]);
  assert.deepEqual(adminLegacyContributions("dev", { manifests: [manifest("base", [target.page])], installations: [installation("base")] }, "files").map(item => item.page.id), ["main"]);
  assert.deepEqual(adminLegacyContributions("dev", { manifests: [manifest("base", [target.page])], installations: [installation("base")] }, "database"), []);
});
