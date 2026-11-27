import assert from "node:assert/strict";
import { test } from "node:test";
import { decodePlaygroundResult, preparePlayground, routeIdentity, searchRoutes } from "../../admin-nextjs/app/lib/admin-playground.js";
import type { ProjectHTTPRoute } from "../../admin-nextjs/app/lib/admin-types.js";

const route: ProjectHTTPRoute = { plugin_id: "files", version: "1", operation: "files.metadata", path: "/alpha/files/metadata", method: "GET", auth: "operator", capability: "files.metadata", function: "metadata" };
const other: ProjectHTTPRoute = { ...route, version: "2", method: "POST", auth: "user", path: "/alpha/files/write" };
test("route filters preserve exact version and authentication", () => {
  assert.equal(routeIdentity(route), "files:1:files.metadata");
  assert.deepEqual(searchRoutes([route, other], " FILES ", "", ""), [route, other]);
  assert.deepEqual(searchRoutes([route, other], "metadata", "GET", "operator"), [route]);
  assert.deepEqual(searchRoutes([route, other], "write", "", "user"), [other]);
  assert.deepEqual(searchRoutes([route, other], "absent", "", ""), []);
});
test("confirmation freezes a valid bounded command for the discovered version", () => {
  const command = preparePlayground("alpha", [route, other], routeIdentity(other), '{"key":"abc"}');
  assert.deepEqual(JSON.parse(command.body), { operation: "plugin.invoke", project_id: "alpha", plugin_id: "files", version: "2", target_operation: "files.metadata", payload: { key: "abc" } });
  assert.throws(() => preparePlayground("", [route], routeIdentity(route), "{}"), /active project/);
  assert.throws(() => preparePlayground("alpha", [route], "", "{}"), /Select a route/);
  assert.throws(() => preparePlayground("alpha", [route], routeIdentity(other), "{}"), /no longer available/);
  assert.throws(() => preparePlayground("alpha", [route], routeIdentity(route), "{"), /Invalid JSON/);
  assert.throws(() => preparePlayground("alpha", [route], routeIdentity(route), JSON.stringify("é".repeat(40000))), /64 KiB/);
  assert.deepEqual(JSON.parse(preparePlayground("alpha", [route], routeIdentity(route), "").body).payload, {});
});
test("response status cannot be inferred from transport success; decode Unicode safely", () => {
  const encoded = Buffer.from(JSON.stringify({ value: "سلام" })).toString("base64");
  assert.equal(decodePlaygroundResult({ Success: true, Payload: encoded }).output, JSON.stringify({ value: "سلام" }, null, 2));
  assert.equal(decodePlaygroundResult({ success: false, errorCode: "denied" }).message, "Plugin error: denied");
  assert.equal(decodePlaygroundResult({ Success: false }).success, false);
  assert.equal(decodePlaygroundResult({ Success: true, Payload: btoa("plain") }).output, '"plain"');
  assert.equal(decodePlaygroundResult({ Success: true, Payload: "!!!" }).output, '"!!!"');
  assert.throws(() => decodePlaygroundResult(null), /Invalid host/);
  assert.throws(() => decodePlaygroundResult({}), /Missing worker/);
});
