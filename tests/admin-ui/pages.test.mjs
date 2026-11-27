import assert from "node:assert/strict";
import { test } from "node:test";
import { app, mount, find, button, click, change, text, session, stubFetch, json, navigations } from "./dom-harness.mjs";

const Functions = app("functions/view.tsx").default;
const FunctionRunner = app("functions/function-runner.tsx").FunctionRunner;
const Workers = app("workers/view.tsx").default;
const Files = app("components/legacy-admin-page.tsx").default;

test("control-plane route modules use the shared page implementation", () => {
  const main = app("view.tsx").default;
  assert.equal(app("%5Fsystem/ui/page.tsx").default, main);
  assert.equal(app("[project]/ui/page.tsx").default, main);
  for (const route of ["audit", "projects"]) assert.equal(app("%5Fsystem/ui/" + route + "/page.tsx").default, main, route);
  for (const route of ["system", "routes", "plugins", "monitoring"]) assert.equal(app("[project]/ui/" + route + "/page.tsx").default, main, route);
  assert.equal(app("[project]/ui/[legacy]/page.tsx").default, Files);
  for (const [route, view] of [["dashboard", "dashboard"], ["workers", "workers"], ["functions", "functions"], ["functions/new", "functions/new"], ["functions/[name]", "functions/[name]"], ["plugins/[plugin]", "plugins/[plugin]"]]) assert.equal(app("[project]/ui/" + route + "/page.tsx").default, app(view + "/view.tsx").default);
  assert.equal(app("[project]/ui/extensions/[plugin]/[version]/[page]/page.tsx").default, app("extensions/[project]/[plugin]/[version]/[page]/view.tsx").default);
});

test("legacy resource pages handle absent projects, denied catalog and manifest contributions", async () => {
  await mount(Files, {}, { pathname: "/dev/ui/files" });
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
    if (url === "/_system/admin/projects") {
      if (mode === "projects-denied") return json({}, 403);
      return json([{ ID: "dev", Active: mode !== "inactive" }]);
    }
    if (url === "/dev/admin/plugins") {
      if (mode === "catalog-denied") return json({}, 403);
      return json(mode === "empty" ? { manifests: [], installations: [] } : catalog);
    }
    throw new Error("unexpected request " + url);
  });
  await mount(Files, {}, { pathname: "/dev/ui/files" });
  assert.match(text(), /Projects unavailable/);
  mode = "inactive";
  await mount(Files, {}, { pathname: "/dev/ui/files" });
  assert.match(text(), /Select an active project/);
  mode = "catalog-denied";
  await mount(Files, {}, { pathname: "/dev/ui/files" });
  assert.match(text(), /Plugin catalog unavailable/);
  mode = "empty";
  await mount(Files, {}, { pathname: "/dev/ui/files" });
  assert.match(text(), /Page unavailable/);
  assert.equal(find('a[href="/dev/ui/plugins"]').textContent, "Manage plugins");
  mode = "enabled";
  await mount(Files, {}, { pathname: "/dev/ui/files" });
  assert.match(text(), /Read object/);
  assert.ok(calls.every(call => call.auth === "Bearer fixture-token"));
  await mount(Files, {}, { pathname: "/dev/ui/database" });
  assert.match(text(), /Page unavailable/);
});

test("function list renders public and private versions only after plugin access succeeds", async () => {
  await mount(Functions, {}, { pathname: "/dev/ui/functions" });
  assert.deepEqual(navigations.at(-1), ["replace", "/_system/ui"]);
  session();
  let enabled = false;
  let drafts = [];
  stubFetch(url => {
    if (url === "/_system/admin/projects") return json([{ ID: "dev", Active: true }]);
    if (url.endsWith("/admin/plugins")) return json({ manifests: [{ ID: "functions", Version: "1", Capabilities: [] }], installations: [{ ProjectID: "dev", PluginID: "functions", Version: "1", Enabled: enabled, Ready: true }] });
    if (url === "/dev/admin/functions") return json(drafts);
    throw new Error("unexpected request " + url);
  });
  await mount(Functions, {}, { pathname: "/dev/ui/functions" });
  assert.doesNotMatch(text(), /Create function/);
  enabled = true;
  await mount(Functions, {}, { pathname: "/dev/ui/functions" });
  assert.match(text(), /No functions yet/);
  drafts = [
    { project_id: "dev", name: "public echo", source: "", version: 2, public: true, created_at: "2026-01-01T00:00:00Z" },
    { project_id: "dev", name: "private", source: "", created_at: "2026-01-01T00:00:00Z" },
  ];
  await mount(Functions, {}, { pathname: "/dev/ui/functions" });
  assert.match(text(), /Public/);
  assert.match(text(), /Private/);
  assert.match(text(), /Published records/);
  assert.match(text(), /Last published/);
  assert.doesNotMatch(text(), /Updated/);
  assert.equal(find('a[href="/dev/ui/functions/public%20echo"]').textContent, "public echo");
  assert.doesNotMatch(text(), /No functions yet/);
  await change(find('input[type="search"]'), "private");
  assert.equal(document.querySelector('a[href="/dev/ui/functions/public%20echo"]'), null);
  assert.ok(find('a[href="/dev/ui/functions/private"]'));
  await change(find('input[type="search"]'), "missing");
  assert.match(text(), /No functions match the filters/);
  await change(find('input[type="search"]'), "");
  await click(button("Refresh functions"));
  assert.ok(find('a[href="/dev/ui/functions/private"]'));
});

test("workers page sends a scoped restart request and reloads readiness", async () => {
  await mount(Workers, {}, { pathname: "/dev/ui/workers" });
  assert.deepEqual(navigations.at(-1), ["replace", "/_system/ui"]);
  session();
  let active = true;
  let ready = false;
  let denied = true;
  const calls = [];
  stubFetch((url, init) => {
    if (url === "/_system/admin/projects") return json([{ ID: "dev", Active: active }]);
    if (url === "/dev/admin/plugins") return json({ installations: [
      { ProjectID: "dev", PluginID: "rooms", Version: "1", Enabled: true, Ready: true, WorkerReady: ready },
      { ProjectID: "dev", PluginID: "disabled", Version: "1", Enabled: false, Ready: true, WorkerReady: false },
      { ProjectID: "dev", PluginID: "pending", Version: "1", Enabled: true, Ready: false, WorkerReady: false },
    ] });
    if (url === "/dev/admin/actions") {
      calls.push({ auth: init.headers.authorization, body: JSON.parse(init.body) });
      if (denied) return new Response("worker denied", { status: 403 });
      ready = true;
      return json({});
    }
    throw new Error("unexpected request " + url);
  });
  active = false;
  await mount(Workers, {}, { pathname: "/dev/ui/workers" });
  assert.match(text(), /Project: —/);
  active = true;
  await mount(Workers, {}, { pathname: "/dev/ui/workers" });
  await click(button("Start worker"));
  await click(button("Confirm start"));
  assert.match(text(), /worker denied/);
  denied = false;
  await click(button("Confirm start"));
  assert.match(text(), /Worker started/);
  assert.match(text(), /Worker ready/);
  await click(button("Close"));
  assert.ok(button("Refresh workers"));
  assert.ok(![...document.querySelectorAll("button")].some(node => node.textContent === "Start worker"));
  assert.deepEqual(calls.at(-1), { auth: "Bearer fixture-token", body: { operation: "plugin.worker.start", project_id: "dev", plugin_id: "rooms", version: "1" } });
});

test("plugin details stage a new version without changing routing, and confirm new capability approval", async () => {
  const Plugin = app("plugins/[plugin]/view.tsx").default;
  await mount(Plugin, {}, { pathname: "/dev/ui/plugins/sample%402", params: { plugin: "sample%402" } });
  assert.deepEqual(navigations.at(-1), ["replace", "/_system/ui"]);
  session();
  const first = { ID: "sample", Title: "Sample", Description: "First", Version: "1", Capabilities: ["read"], ConfigSchema: {}, Routes: [] };
  const candidate = { ...first, Version: "2", Capabilities: ["read", "write"] };
  const installations = [{ ProjectID: "dev", PluginID: "sample", Version: "1", Enabled: true, Ready: true, WorkerReady: true, Draining: false, Promoted: true, Grants: ["read"] }];
  const calls = [];
  stubFetch((url, init) => {
    if (url === "/_system/admin/projects") return json([{ ID: "dev", Active: true }]);
    if (url === "/dev/admin/plugins") return json({ manifests: [first, candidate], installations });
    if (url === "/dev/admin/actions" || url === "/_system/admin/actions") {
      const body = JSON.parse(init.body); calls.push({ url, body });
      const current = installations.find(item => item.Version === "2");
      if (body.operation === "plugin.install") installations.push({ ProjectID: "dev", PluginID: "sample", Version: "2", Ready: false, Enabled: false, Draining: false, Grants: [] });
      else if (body.operation === "plugin.readiness") current.Ready = true;
      else if (body.operation === "plugin.grant") current.Grants.push(body.capability);
      else if (body.operation === "plugin.approve-capability") current.ApprovedCapabilities = [body.capability];
      else if (body.operation === "plugin.enable") { current.Enabled = true; current.WorkerReady = true; }
      else if (body.operation === "plugin.promote") { installations[0].Promoted = false; current.Promoted = true; }
      else throw new Error("unexpected operation " + body.operation);
      return new Response(null, { status: 204 });
    }
    throw new Error("unexpected request " + url);
  });
  await mount(Plugin, {}, { pathname: "/dev/ui/plugins/sample%402", params: { plugin: "sample%402" } });
  assert.match(text(), /Promoted version: 1/);
  assert.ok(button("Stage installation"));
  assert.equal([...document.querySelectorAll("button")].some(node => node.textContent === "Install"), false);
  await click(button("Stage installation"));
  assert.equal(calls.length, 0);
  await click(button("Confirm action"));
  assert.equal(calls.at(-1).body.operation, "plugin.install");
  assert.equal(calls.at(-1).body.confirmed, true);
  await click(button("Close"));
  assert.match(text(), /New capability approvals required: write/);
  await click(button("Mark artifact ready"));
  await click(button("Confirm action"));
  assert.equal(calls.at(-1).url, "/_system/admin/actions");
  await click(button("Close"));
  assert.equal(button("Enable").disabled, true);
  for (const capability of ["read", "write"]) {
    const row = [...document.querySelectorAll("tr")].find(row => row.querySelector("td code")?.textContent === capability);
    await click(button("Grant", row));
    await click(button("Confirm action"));
    assert.equal(calls.at(-1).body.capability, capability);
    await click(button("Close"));
  }
  await click(button("Enable"));
  await click(button("Confirm action"));
  await click(button("Close"));
  assert.equal([...document.querySelectorAll("button")].some(node => node.textContent === "Promote version"), false);
  await click(button("Approve upgrade"));
  await click(button("Confirm action"));
  assert.equal(calls.at(-1).body.operation, "plugin.approve-capability");
  assert.equal(calls.at(-1).body.capability, "write");
  await click(button("Close"));
  await click(button("Promote version"));
  await click(button("Confirm action"));
  await click(button("Close"));
  assert.match(text(), /Promoted version: 2/);
  assert.ok(calls.every(call => call.body.project_id === "dev" && call.body.version === "2"));
});

test("sandbox runner rejects invalid JSON and sends the current source in a scoped operator request", async () => {
  session();
  const calls = [];
  stubFetch((url, init) => { calls.push({ url, body: JSON.parse(init.body), headers: init.headers }); return json({ valid: true, result: { ok: true } }); });
  const source = "export default async function handler(input) { return input; }";
  await mount(FunctionRunner, { project: "dev", name: "echo", source });
  await change(find("textarea"), "{");
  await click(button("Run test"));
  assert.match(text(), /Input must be valid JSON/);
  assert.equal(calls.length, 0);
  await change(find("textarea"), '{"value":42}');
  await click(button("Run test"));
  assert.equal(calls.length, 1);
  assert.equal(calls[0].url, "/dev/admin/functions");
  assert.equal(calls[0].headers.authorization, "Bearer fixture-token");
  assert.deepEqual(calls[0].body, { operation: "functions.run", project_id: "dev", name: "echo", source, input: { value: 42 } });
  assert.match(text(), /Execution succeeded/);
  assert.match(text(), /"ok": true/);
  await change(find("textarea"), "null");
  assert.match(text(), /previous source or input/);
});
