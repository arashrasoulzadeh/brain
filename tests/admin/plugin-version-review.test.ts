import assert from "node:assert/strict";
import { test } from "node:test";
import { pluginVersionReview, pluginCatalogState } from "../../admin-nextjs/app/lib/plugin-version-review.js";
import type { PluginManifest, PluginStatus } from "../../admin-nextjs/app/lib/admin-types.js";

const one: PluginManifest = { ID: "sample", Version: "1", Title: "Sample", Description: "", Capabilities: ["read"], ConfigSchema: { name: "string" } };
const two: PluginManifest = { ...one, Version: "2", Capabilities: ["read", "write"], ConfigSchema: { name: "bool", added: "int" }, Dependencies: ["dep@1"], OptionalDependencies: ["missing@1"] };
const dep: PluginManifest = { ID: "dep", Version: "1", Title: "Dependency", Description: "", Capabilities: ["use"] };
function installation(id: string, version: string, extra: Partial<PluginStatus> = {}): PluginStatus {
  return { ProjectID: "alpha", PluginID: id, Version: version, Enabled: true, Ready: true, Draining: false, WorkerReady: true, Grants: ["read"], ...extra };
}
test("version review is project scoped and distinguishes approval, config and exact dependency state", () => {
  const first = installation("sample", "1", { Promoted: true, ConfiguredFields: ["name"] });
  const second = installation("sample", "2", { Enabled: false });
  const foreign = installation("dep", "1", { ProjectID: "beta", Grants: ["use"] });
  const review = pluginVersionReview("alpha", two, [one, two, dep], [first, second, foreign]);
  assert.equal(review.promoted?.manifest.Version, "1");
  assert.deepEqual(review.missingApprovals, ["write"]);
  assert.deepEqual(review.missingGrants, ["write"]);
  assert.deepEqual(review.missingConfig, ["name", "added"]);
  assert.deepEqual(review.incompatibleConfig, ["name"]);
  assert.deepEqual(review.requiredUnavailable.map(item => item.key), ["dep@1"]);
  assert.equal(review.dependencies[1].registered, false);
  const updated = installation("sample", "2", { Grants: ["read", "write"], ApprovedCapabilities: ["write"], ConfiguredFields: ["name", "added"] });
  const readyDep = installation("dep", "1", { Grants: ["use"] });
  const ready = pluginVersionReview("alpha", { ...two, ConfigSchema: { name: "string", added: "int" } }, [one, two, dep], [first, updated, readyDep]);
  assert.deepEqual(ready.missingApprovals, []);
  assert.deepEqual(ready.missingGrants, []);
  assert.deepEqual(ready.incompatibleConfig, []);
  assert.deepEqual(ready.requiredUnavailable, []);
});
test("catalog states reflect readiness, draining and worker availability", () => {
  assert.equal(pluginCatalogState(undefined), "uninstalled");
  assert.equal(pluginCatalogState(installation("sample", "1", { Draining: true })), "draining");
  assert.equal(pluginCatalogState(installation("sample", "1", { Ready: false })), "not-ready");
  assert.equal(pluginCatalogState(installation("sample", "1", { Enabled: false })), "disabled");
  assert.equal(pluginCatalogState(installation("sample", "1", { WorkerReady: false })), "worker-unavailable");
  assert.equal(pluginCatalogState(installation("sample", "1")), "enabled");
});
