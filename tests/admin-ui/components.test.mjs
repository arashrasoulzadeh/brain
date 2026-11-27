import assert from "node:assert/strict";
import { createRequire } from "node:module";
import { test } from "node:test";

const require = createRequire(new URL("../../admin-nextjs/package.json", import.meta.url));
const { createElement: h } = require("react");
const { renderToStaticMarkup: render } = require("react-dom/server");
const { AdminResultRenderer } = require("./app/components/admin-results.tsx");
const { AuditPanel } = require("./app/components/audit-panel.tsx");
const { MonitoringPanel } = require("./app/components/monitoring-panel.tsx");
const { SystemPanel } = require("./app/components/system-panel.tsx");
const { ProjectsPanel } = require("./app/components/projects-panel.tsx");
const { RoutesPanel } = require("./app/components/routes-panel.tsx");
const { AdminConfigForm } = require("./app/components/admin-config-form.tsx");
const { copy } = require("./app/lib/admin-copy.ts");
const text = copy.en;
const noop = () => {};
const html = (component, props) => render(h(component, props));

test("result renderer supports every documented view and absent data", () => {
  const cases = [
    [[{ name: "Alpha", amount: 2 }, { name: "Beta", amount: -1 }], { type: "table", columns: ["name"], page_size: 1 }, /Alpha/],
    [[], { type: "table" }, /No rows/],
    [42, { type: "table" }, /42/],
    [{ count: 4 }, { type: "detail", title: "Stats" }, /<dt>count<\/dt><dd>4<\/dd>/],
    [42, { type: "detail" }, /42/],
    [12, { type: "metric" }, /<strong[^>]*>12<\/strong>/],
    [[{ name: "Alpha", amount: 2 }, { name: "Beta", amount: -1 }, { name: "Invalid", amount: "x" }], { type: "chart", label_field: "name", value_field: "amount" }, /Beta · -1/],
    [null, { type: "chart" }, /No numeric chart data/],
    [["first", "second"], { type: "logs" }, /first\s+second/],
    ["line", { type: "logs" }, /line/],
    ["source", { type: "code" }, /source/],
    [42, { type: "grid", children: [{ type: "metric" }] }, /42/],
    [42, { type: "tabs", children: [{ title: "Count", type: "metric" }] }, /role="tabpanel"/],
    [42, { type: "tabs" }, /role="tablist"/],
    [{ value: "hi" }, { type: "json" }, /hi/],
    [{}, { path: "absent" }, /No result for absent/],
    [undefined, undefined, /No result for this component/],
    [[{ id: "row" }], undefined, /row/],
    [{ total: 1, rows: [{ id: "row" }] }, undefined, /row/],
    ["plain", undefined, /plain/],
  ];
  for (const [value, spec, expected] of cases) assert.match(html(AdminResultRenderer, { value, spec }), expected);
});

test("monitoring distinguishes every lifecycle state", () => {
  const states = [
    ["absent", null, "not installed"],
    ["draining", { Draining: true, Ready: true, Enabled: true }, "draining"],
    ["pending", { Ready: false, Enabled: true }, text.notReady],
    ["disabled", { Ready: true, Enabled: false }, text.disabled],
    ["unavailable", { Ready: true, Enabled: true, WorkerReady: false }, "worker unavailable"],
    ["healthy", { Ready: true, Enabled: true, WorkerReady: true }, text.ready],
  ];
  const manifests = states.map(([ID]) => ({ ID, Version: "1" }));
  const installations = states.flatMap(([PluginID, state]) => state ? [{ PluginID, Version: "1", ...state }] : []);
  const markup = html(MonitoringPanel, { activePage: "monitoring", text, manifests, installations });
  for (const [id, , label] of states) assert.ok(markup.includes(id + "@1 · " + label), markup);
  assert.match(markup, /class="success"/);
  assert.match(html(MonitoringPanel, { activePage: "projects", text, manifests: [], installations: [] }), /page-hidden/);
});

test("audit filtering includes project, source and detail without reordering input", () => {
  const audit = [
    { action: "grant", source: "operator", project_id: "one", plugin_id: "rooms", version: "1", detail: "allowed", success: true },
    { action: "deny", source: "host", success: false },
  ];
  const props = { activePage: "audit", text, audit, auditFilter: "", setAuditFilter: noop };
  const full = html(AuditPanel, props);
  assert.ok(full.indexOf("deny") < full.indexOf("grant"));
  assert.match(full, /class="failure"/);
  assert.match(full, /class="success"/);
  assert.match(html(AuditPanel, { ...props, auditFilter: "ALLOWED" }), /grant/);
  assert.doesNotMatch(html(AuditPanel, { ...props, auditFilter: "ALLOWED" }), /<strong>deny/);
  assert.match(html(AuditPanel, { ...props, activePage: "projects", audit: [] }), /No audit events yet/);
  assert.equal(audit[0].action, "grant");
});

test("project and route screens render scope and inactive project restrictions", () => {
  const props = { activePage: "projects", text, selectedProjectID: "one", projectID: "", projects: [{ ID: "one", Active: true }, { ID: "two", Active: false }], setProjectID: noop, createProject: noop, setProjectActive: noop, uninstallProject: noop, onSelectProject: noop, error: "fixture error" };
  const markup = html(ProjectsPanel, props);
  assert.match(markup, /two \(inactive\)/);
  assert.match(markup, /fixture error/);
  assert.match(markup, /Deactivate/);
  assert.match(markup, /Activate/);
  assert.match(html(ProjectsPanel, { ...props, activePage: "system", projects: [], error: "" }), /No projects yet/);
  const system = [{ ID: "rooms", Title: "Rooms", Version: "1", Routes: [{ Operation: "rooms.list", Capability: "rooms.read" }], Capabilities: ["rooms.read"], ExtensionPoints: [], SourceFiles: ["index.ts"], Imports: ["sdk"], Dependencies: ["auth@1"], PluginImports: ["tables@1"] }, { ID: "empty", Version: "1" }];
  const routeProps = { activePage: "routes", text, system, playgroundRoute: "rooms:rooms.list", playgroundPayload: "{}", selectedProjectID: "one", setPlaygroundRoute: noop, setPlaygroundPayload: noop, testPlayground: noop, sendPlayground: noop, playgroundMessage: "Preview ready" };
  assert.match(html(RoutesPanel, routeProps), /rooms.list/);
  assert.match(html(RoutesPanel, routeProps), /project_id/);
  assert.match(html(RoutesPanel, { ...routeProps, activePage: "system", playgroundRoute: "" }), /Select a route/);
  const inventory = html(SystemPanel, { activePage: "system", text, system });
  assert.match(inventory, /rooms.read/);
  assert.match(inventory, /auth@1 · tables@1/);
  assert.match(html(SystemPanel, { activePage: "projects", text, system: [] }), /page-hidden/);
});

test("configuration form renders only declared editable fields", () => {
  const props = { token: "fixture", projectID: "p", pluginID: "rooms", version: "1", schema: { enabled: "bool", limit: "int", title: "string", unsupported: "secret" } };
  const markup = html(AdminConfigForm, props);
  assert.match(markup, /<select/);
  assert.match(markup, /type="number"/);
  assert.match(markup, /unsupported \(secret\)/);
  assert.match(markup, /Save configuration/);
  assert.equal(html(AdminConfigForm, { ...props, schema: {} }), "");
});
