import assert from "node:assert/strict";
import { test } from "node:test";
import { adminContributions, adminPageExtensions } from "../../admin-nextjs/app/lib/admin-registry.js";
import type { AdminCatalog } from "../../admin-nextjs/app/lib/admin-registry.js";

test("admin contributions require exact project, version, readiness and grants", () => {
  const catalog: AdminCatalog = {
    manifests: [{ ID: "custom", Title: "Custom", Description: "", Version: "1", Capabilities: ["read"], Routes: [{ Operation: "custom.list", Capability: "read" }], ExtensionPoints: [], SourceFiles: [], Imports: [], PluginImports: [], Dependencies: [], Admin: { version: 1, pages: [{ id: "list", title: "List", actions: [{ title: "List", operation: "custom.list" }] }] } }],
    installations: [{ ProjectID: "one", PluginID: "custom", Version: "1", Enabled: true, Ready: true, Draining: false, Grants: ["read"] }],
  };
  assert.equal(adminContributions("one", catalog)[0].href, "/extensions/one/custom/1/list");
  assert.deepEqual(adminContributions("two", catalog), []);
  const installation = catalog.installations![0];
  for (const change of [{ Version: "2" }, { Ready: false }, { Enabled: false }, { Draining: true }, { Grants: [] }]) {
    assert.deepEqual(adminContributions("one", { ...catalog, installations: [{ ...installation, ...change }] }), []);
  }
});

test("cross-plugin contributions remain project scoped", () => {
  const page = { id: "list", title: "Rooms", actions: [] };
  const target = { projectID: "one", pluginID: "rooms", version: "1", page, href: "/rooms" };
  const extension = { ...target, pluginID: "custom", href: "/custom", page: { ...page, extends: "rooms/list" } };
  assert.deepEqual(adminPageExtensions(target, [extension, { ...extension, projectID: "two" }]), [extension]);
});
