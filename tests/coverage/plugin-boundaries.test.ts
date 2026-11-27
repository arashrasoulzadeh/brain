import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import { test } from "node:test";
import { describe } from "../../plugins/access-management/index.js";
import { example, invoke, observeRoomConnected } from "../../plugins/template/index.js";
import type { PluginContext } from "../../plugins/sdk-typescript/src/index.js";
import { adminAction, adminLogin } from "../../clients/javascript/src/admin.js";

test("plugin descriptions, requests and observations respect project scope", () => {
  const events: unknown[] = [];
  const context: PluginContext = {
    project: { id: "p1", active: true },
    plugin: { id: "template", version: "1" },
    execution: { operation: "example" },
    plugins: { observe: { fire(event, payload) { events.push({ event, payload }); } } },
  };
  assert.deepEqual(describe(context), { projectId: "p1", policy: "deny-by-default", scope: "invocation" });
  assert.equal(invoke("p1", "template", "1", "example").params.projectId, "p1");
  assert.deepEqual(example(context, {}), { projectId: "p1", accepted: true });
  assert.deepEqual(events, [{ event: "exampleCompleted", payload: {} }]);
  const event = { id: "event1", projectId: "p1", pluginId: "rooms", version: "1", event: "connected", payload: { roomId: "r1", reconnected: false } };
  assert.deepEqual(observeRoomConnected(context, event), { recorded: true });
  for (const invalid of [
    { ...event, projectId: "p2" }, { ...event, pluginId: "other" },
    { ...event, event: "disconnected" }, { ...event, payload: null },
    { ...event, payload: { roomId: 1, reconnected: false } },
    { ...event, payload: { roomId: "r1", reconnected: "false" } },
  ]) assert.throws(() => observeRoomConnected(context, invalid));
});

test("SDK worker handles successful requests, handler errors, malformed and oversized input", () => {
  const request = (id: string) => JSON.stringify({ jsonrpc: "2.0", id, method: "invoke", params: { protocol: 1, projectId: "p1", pluginId: "template", version: "1", operation: "example", payload: { value: 42 } } });
  const result = spawnSync(process.execPath, ["--import", "tsx", "--input-type=module", "-e",
    'import { startWorker } from "./plugins/sdk-typescript/src/worker.ts"; startWorker(request => { if (request.id === "fail") throw new Error("test"); return request.params.payload; });',
  ], { input: [request("ok"), request("fail"), "{", "x".repeat(1024 * 1024 + 1), ""].join("\n"), encoding: "utf8", timeout: 10000 });
  assert.equal(result.status, 0, result.stderr);
  const responses = result.stdout.trim().split("\n").map(line => JSON.parse(line));
  assert.equal(responses.length, 4);
  assert.deepEqual(responses.find(r => r.id === "ok"), { jsonrpc: "2.0", id: "ok", result: { protocol: 1, success: true, payload: { value: 42 } } });
  assert.equal(responses.find(r => r.id === "fail")?.errorCode, "plugin_error");
  assert.ok(responses.some(r => r.errorCode === "invalid_json"));
  assert.ok(responses.some(r => r.errorCode === "payload_too_large"));
});

test("admin client rejects unsuccessful HTTP responses and supports default transport", async () => {
  const denied: typeof fetch = async () => new Response("", { status: 403 });
  await assert.rejects(adminLogin("https://example.test", "user", "secret", denied), /admin login failed/);
  await assert.rejects(adminAction("https://example.test", "token", { operation: "list" }, denied), /admin action failed/);
  const previous = globalThis.fetch;
  try {
    globalThis.fetch = async (input, init) => {
      if (String(input).endsWith("/admin/session")) {
        assert.equal(init?.method, "POST");
        return Response.json({ session: "token" });
      }
      assert.equal(new Headers(init?.headers).get("authorization"), "Bearer token");
      return new Response(null, { status: 204 });
    };
    assert.deepEqual(await adminLogin("https://example.test", "user", "secret"), { session: "token" });
    await adminAction("https://example.test", "token", { operation: "list" });
  } finally { globalThis.fetch = previous; }
});
