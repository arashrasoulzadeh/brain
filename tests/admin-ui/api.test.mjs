import assert from "node:assert/strict";
import { createRequire } from "node:module";
import { test } from "node:test";

const require = createRequire(new URL("../../admin-nextjs/package.json", import.meta.url));
const api = require("./app/lib/admin-api.ts");
const sdk = require("./app/lib/admin-sdk.ts");
const contribution = { projectID: "p", pluginID: "rooms", version: "1", page: { id: "main", actions: [{ operation: "rooms.list", read_only: true }, { operation: "rooms.create", read_only: false }] } };

test("session storage and sign-out preserve operator boundaries", async () => {
  const previousWindow = globalThis.window;
  const previousFetch = globalThis.fetch;
  const values = new Map();
  const events = [];
  globalThis.window = { dispatchEvent: event => { events.push(event.type); return true; }, sessionStorage: { getItem: key => values.get(key) ?? null, setItem: (key, value) => values.set(key, value), removeItem: key => values.delete(key) } };
  try {
    assert.equal(api.readAdminSession(), "");
    api.saveAdminSession("fixture");
    assert.equal(api.readAdminSession(), "fixture");
    const projects = [{ ID: "a", Active: false }, { ID: "b", Active: true }, { ID: "c", Active: true }];
    assert.equal(api.preferredAdminProject(projects).ID, "b");
    api.saveAdminProject("c");
    assert.equal(api.preferredAdminProject(projects).ID, "c");
    api.saveAdminProject("a");
    assert.equal(api.preferredAdminProject(projects), undefined);
    api.saveAdminProject("");
    assert.deepEqual(events, [api.ADMIN_SESSION_CHANGED, ...Array(3).fill(api.ADMIN_PROJECT_CHANGED)]);
    assert.equal(api.preferredAdminProject(projects).ID, "b");
    globalThis.fetch = async (url, init) => {
      assert.equal(url, "/_system/admin/session");
      assert.equal(init.method, "DELETE");
      assert.equal(init.headers.authorization, "Bearer fixture");
      throw new Error("offline");
    };
    await api.signOutAdminSession();
    assert.equal(api.readAdminSession(), "");
    assert.equal(values.size, 0);
    await api.revokeAdminSession("");
  } finally { globalThis.window = previousWindow; globalThis.fetch = previousFetch; }
});

test("admin SDK validates action declarations, project scope and response envelopes", async () => {
  const previous = globalThis.fetch;
  const signal = new AbortController().signal;
  let next = Response.json({ Success: true, Payload: Buffer.from(JSON.stringify({ value: "é" })).toString("base64") });
  const calls = [];
  globalThis.fetch = async (url, init) => { calls.push({ url, init, body: JSON.parse(init.body) }); return next; };
  try {
    await assert.rejects(sdk.invokeAdminAction("token", contribution, "unknown", {}, signal), /Action unavailable/);
    await assert.rejects(sdk.invokeAdminAction("token", contribution, "rooms.create", {}, signal, true), /Action unavailable/);
    assert.equal(calls.length, 0);
    assert.deepEqual(await sdk.invokeAdminAction("token", contribution, "rooms.list", {}, signal, true), { value: "é" });
    assert.deepEqual(calls[0].body, { operation: "plugin.admin.query", project_id: "p", plugin_id: "rooms", version: "1", page_id: "main", target_operation: "rooms.list", payload: {}, confirmed: false });
    assert.equal(calls[0].init.signal, signal);
    assert.equal(calls[0].init.headers.authorization, "Bearer token");
    next = Response.json({ Success: true });
    assert.equal(await sdk.invokeAdminAction("token", contribution, "rooms.create", { id: "one" }, signal), null);
    assert.equal(calls[1].body.confirmed, true);
    for (const [response, error] of [[new Response("denied", { status: 403 }), /denied/], [Response.json({ Success: false, ErrorCode: "Forbidden" }), /Forbidden/], [Response.json({ Success: false }), /Plugin operation failed/]]) {
      next = response;
      await assert.rejects(sdk.invokeAdminAction("token", contribution, "rooms.list", {}, signal), error);
    }
    next = Response.json({ revision: 3 });
    assert.equal(await sdk.adminPageRevision("token", contribution, signal), 3);
    next = Response.json({ revision: 1.5 });
    await assert.rejects(sdk.adminPageRevision("token", contribution, signal), /Invalid observation revision/);
    next = new Response("denied", { status: 403 });
    await assert.rejects(sdk.adminPageRevision("token", contribution, signal), /denied/);
    next = Response.json({ rows: [{ key: "a" }], next: "a" });
    assert.deepEqual(await sdk.fetchAdminInventory("token", contribution, "", signal), { rows: [{ key: "a" }], next: "a" });
    assert.equal(calls.at(-1).body.limit, 50);
    next = Response.json({ rows: [] });
    assert.deepEqual(await sdk.fetchAdminInventory("token", contribution, "a", signal), { rows: [] });
    for (const bad of [{ rows: null }, { rows: [{ key: 1 }] }, { rows: [], next: 1 }]) {
      next = Response.json(bad);
      await assert.rejects(sdk.fetchAdminInventory("token", contribution, "", signal), /Invalid inventory response/);
    }
    next = new Response("denied", { status: 403 });
    await assert.rejects(sdk.fetchAdminInventory("token", contribution, "", signal), /denied/);
  } finally { globalThis.fetch = previous; }
});
