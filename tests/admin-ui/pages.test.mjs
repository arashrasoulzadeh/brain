import assert from "node:assert/strict";
import { test } from "node:test";
import { app, mount, find, button, click, text, session, stubFetch, json, navigations } from "./dom-harness.mjs";

const Functions = app("functions/page.tsx").default;
const Workers = app("workers/page.tsx").default;
const Files = app("files/page.tsx").default;

test("control-plane route modules use the shared page implementation", () => {
  const main = app("page.tsx").default;
  for (const route of ["audit", "system", "routes", "plugins", "projects", "monitoring"]) {
    assert.equal(app(route + "/page.tsx").default, main, route);
  }
  for (const route of ["database", "live-sessions", "workflows"]) {
    assert.equal(app(route + "/page.tsx").default, Files, route);
  }
});

test("legacy resource pages handle absent projects, denied catalog and manifest contributions", async () => {
  await mount(Files, {}, { pathname: "/files" });
  assert.match(text(), /Sign in to continue/);
  session();
  let mode = "projects-denied";
  const calls = [];
  const catalog = {
    manifests: [{ ID: "files", Version: "1", Capabilities: ["files.read"],
      Routes: [{ Operation: "files.read", Capability: "files.read" }],
      Admin: { version: 1, pages: [{ id: "read", title: "Objects", legacy_route: "files", actions: [{ operation: "files.read", title: "Read object", read_only: true, fields: [] }] }] } }],
    installations: [{ ProjectID: "dev", PluginID: "files", Version: "1", Enabled: true, Ready: true, Grants: ["files.read"] }],
  };
  stubFetch((url, init) => {
    calls.push({ url, auth: init.headers?.authorization });
    if (url === "/api/admin/projects") {
      if (mode === "projects-denied") return json({}, 403);
      return json([{ ID: "dev", Active: mode !== "inactive" }]);
    }
    if (url === "/api/admin/plugins?project_id=dev") {
      if (mode === "catalog-denied") return json({}, 403);
      return json(mode === "empty" ? { manifests: [], installations: [] } : catalog);
    }
    throw new Error("unexpected request " + url);
  });
  await mount(Files, {}, { pathname: "/files" });
  assert.match(text(), /Projects unavailable/);
  mode = "inactive";
  await mount(Files, {}, { pathname: "/files" });
  assert.match(text(), /Select an active project/);
  mode = "catalog-denied";
  await mount(Files, {}, { pathname: "/files" });
  assert.match(text(), /Plugin catalog unavailable/);
  mode = "empty";
  await mount(Files, {}, { pathname: "/files" });
  assert.match(text(), /Page unavailable/);
  assert.equal(find('a[href="/plugins"]').textContent, "Manage plugins");
  mode = "enabled";
  await mount(Files, {}, { pathname: "/files" });
  assert.match(text(), /Read object/);
  assert.ok(calls.every(call => call.auth === "Bearer fixture-token"));
  await mount(app("database/page.tsx").default, {}, { pathname: "/database" });
  assert.match(text(), /Page unavailable/);
});

test("function list renders public and private versions only after plugin access succeeds", async () => {
  await mount(Functions, {}, { pathname: "/functions" });
  assert.deepEqual(navigations.at(-1), ["replace", "/"]);
  session();
  let enabled = false;
  let drafts = [];
  stubFetch(url => {
    if (url === "/api/admin/projects") return json([{ ID: "dev", Active: true }]);
    if (url.startsWith("/api/admin/plugins")) return json({ installations: [{ PluginID: "functions", Enabled: enabled, Ready: true }] });
    if (url === "/api/admin/functions?project_id=dev") return json(drafts);
    throw new Error("unexpected request " + url);
  });
  await mount(Functions, {}, { pathname: "/functions" });
  assert.doesNotMatch(text(), /Create function/);
  enabled = true;
  await mount(Functions, {}, { pathname: "/functions" });
  assert.match(text(), /No functions yet/);
  drafts = [
    { project_id: "dev", name: "public echo", source: "", version: 2, public: true, created_at: "2026-01-01T00:00:00Z" },
    { project_id: "dev", name: "private", source: "", created_at: "2026-01-01T00:00:00Z" },
  ];
  await mount(Functions, {}, { pathname: "/functions" });
  assert.match(text(), /Public/);
  assert.match(text(), /Private/);
  assert.equal(find('a[href="/functions/public%20echo"]').textContent, "public echo");
  assert.doesNotMatch(text(), /No functions yet/);
});

test("workers page sends a scoped restart request and reloads readiness", async () => {
  await mount(Workers, {}, { pathname: "/workers" });
  assert.deepEqual(navigations.at(-1), ["replace", "/"]);
  session();
  let active = true;
  let ready = false;
  let denied = true;
  const calls = [];
  stubFetch((url, init) => {
    if (url === "/api/admin/projects") return json([{ ID: "dev", Active: active }]);
    if (url === "/api/admin/plugins?project_id=dev") return json({ installations: [
      { PluginID: "rooms", Version: "1", Enabled: true, Ready: true, WorkerReady: ready },
      { PluginID: "disabled", Version: "1", Enabled: false, Ready: true, WorkerReady: false },
      { PluginID: "pending", Version: "1", Enabled: true, Ready: false, WorkerReady: false },
    ] });
    if (url === "/api/admin/actions") {
      calls.push({ auth: init.headers.authorization, body: JSON.parse(init.body) });
      if (denied) return new Response("worker denied", { status: 403 });
      ready = true;
      return json({});
    }
    throw new Error("unexpected request " + url);
  });
  active = false;
  await mount(Workers, {}, { pathname: "/workers" });
  assert.match(text(), /Project: —/);
  active = true;
  await mount(Workers, {}, { pathname: "/workers" });
  assert.equal(document.querySelectorAll("button").length, 1);
  await click(button("Spin up worker"));
  assert.match(text(), /worker denied/);
  denied = false;
  await click(button("Spin up worker"));
  assert.match(text(), /rooms@1 worker started/);
  assert.match(text(), /Worker ready/);
  assert.equal(document.querySelectorAll("button").length, 0);
  assert.deepEqual(calls.at(-1), { auth: "Bearer fixture-token", body: { operation: "plugin.worker.start", project_id: "dev", plugin_id: "rooms", version: "1" } });
});
