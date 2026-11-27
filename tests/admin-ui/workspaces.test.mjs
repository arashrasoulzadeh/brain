import assert from "node:assert/strict";
import { test } from "node:test";
import { app, mount, find, button, click, change, text, session, stubFetch, json, navigations, act, h } from "./dom-harness.mjs";

const Dashboard = app("dashboard/view.tsx").default;
const Extension = app("extensions/[project]/[plugin]/[version]/[page]/view.tsx").default;
const { AdminShell } = app("components/admin-shell.tsx");
const { AdminPageFrame } = app("components/admin-chrome.tsx");
const { ADMIN_PLUGIN_CATALOG_CHANGED, ADMIN_SESSION_CHANGED } = app("lib/admin-api.ts");

test("dashboard reports project-scoped health and refreshes the catalog", async () => {
  session();
  const manifests = ["dashboard", "rooms", "pending"].map(ID => ({ ID, Title: ID, Version: "1", Capabilities: ["read"] }));
  let ready = false;
  let projectAvailable = true;
  let catalogAvailable = true;
  stubFetch(url => {
    if (url === "/_system/admin/projects") return json([{ ID: "dev", Name: "Fixture project", Active: projectAvailable }]);
    if (url === "/dev/admin/plugins") return catalogAvailable ? json({ manifests, installations: [
      { ProjectID: "dev", PluginID: "dashboard", Version: "1", Enabled: true, Ready: true, WorkerReady: true, Grants: ["read"] },
      { ProjectID: "dev", PluginID: "rooms", Version: "1", Enabled: true, Ready: true, WorkerReady: ready, Grants: ["read"] },
      { ProjectID: "other", PluginID: "foreign", Version: "1", Enabled: true, Ready: true, WorkerReady: true, Grants: ["read"] },
    ] }) : json({}, 403);
    throw new Error("unexpected request " + url);
  });
  await mount(Dashboard, {}, { pathname: "/dev/ui/dashboard" });
  assert.match(text(), /Fixture project/);
  assert.match(text(), /2 installed/);
  assert.match(text(), /1 installations need attention/);
  assert.match(text(), /Worker unavailable/);
  assert.doesNotMatch(text(), /foreign/);
  ready = true;
  await click(button("Refresh dashboard"));
  assert.match(text(), /2 healthy/);
  assert.match(text(), /No readiness, grant or worker warnings/);
  catalogAvailable = false;
  await act(async () => window.dispatchEvent(new window.Event(ADMIN_PLUGIN_CATALOG_CHANGED)));
  assert.match(text(), /Plugin catalog unavailable/);
  catalogAvailable = true;
  await click(button("Retry"));
  await click(button("Refresh dashboard"));
  assert.match(text(), /2 healthy/);
  projectAvailable = false;
  await click(button("Refresh dashboard"));
  assert.match(text(), /Project unavailable/);
});

test("manifest extension invalidates previously authorized content after revocation", async () => {
  session();
  let granted = true;
  const manifest = { ID: "example", Title: "Example", Version: "1", Capabilities: ["example.read"], Routes: [{ Operation: "example.read", Capability: "example.read" }], Admin: { version: 1, pages: [{ id: "read", title: "Example data", actions: [{ operation: "example.read", title: "Read data", read_only: true, fields: [] }] }] } };
  stubFetch(url => {
    if (url === "/_system/admin/projects") return json([{ ID: "dev", Active: true }]);
    if (url === "/dev/admin/plugins") return json({ manifests: [manifest], installations: [{ ProjectID: "dev", PluginID: "example", Version: "1", Enabled: true, Ready: true, Grants: granted ? ["example.read"] : [] }] });
    throw new Error("unexpected request " + url);
  });
  const params = { project: "dev", plugin: "example", version: "1", page: "read" };
  await mount(Extension, {}, { pathname: "/dev/ui/extensions/example/1/read", params });
  assert.match(text(), /Read data/);
  granted = false;
  await act(async () => window.dispatchEvent(new window.Event(ADMIN_PLUGIN_CATALOG_CHANGED)));
  assert.match(text(), /Page unavailable or permission revoked/);
  assert.doesNotMatch(text(), /Read data/);
});

test("shared admin shell hides content without a session and provides heading and sign-out", async () => {
  const child = h(AdminPageFrame, { activePage: "projects", title: "Fixture heading" }, h("p", null, "Protected content"));
  await mount(AdminShell, { children: child }, { pathname: "/_system/ui/projects" });
  assert.match(text(), /Checking operator session/);
  assert.doesNotMatch(text(), /Protected content/);
  assert.deepEqual(navigations.at(-1), ["replace", "/_system/ui"]);
  session();
  stubFetch(url => {
    if (url === "/_system/admin/projects") return json([{ ID: "dev", Active: true }]);
    if (url === "/dev/admin/plugins") return json({ manifests: [], installations: [] });
    if (url === "/_system/admin/session") return json({});
    throw new Error("unexpected request " + url);
  });
  await act(async () => window.dispatchEvent(new window.Event(ADMIN_SESSION_CHANGED)));
  assert.match(text(), /Protected content/);
  assert.equal(find("h1").textContent, "Fixture heading");
  await change(find('input[aria-label="Filter navigation"]'), "nonexistent");
  assert.match(text(), /No matching pages/);
  await click(button("Sign out"));
  assert.equal(window.sessionStorage.getItem("brain-admin-session"), null);
  assert.doesNotMatch(text(), /Protected content/);
  await mount(AdminShell, { children: h("p", null, "Login form") }, { pathname: "/_system/ui" });
  assert.match(text(), /Login form/);
});
