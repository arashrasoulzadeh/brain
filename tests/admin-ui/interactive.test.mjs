import assert from "node:assert/strict";
import { test } from "node:test";
import { app, mount, find, button, click, change, submit, text, session, stubFetch, json, reply, navigations } from "./dom-harness.mjs";

const { AdminResultRenderer } = app("components/admin-results.tsx");
const { AdminConfigForm } = app("components/admin-config-form.tsx");
const { AdminPageRenderer } = app("components/admin-renderer.tsx");
const { AdminInventory } = app("components/admin-inventory.tsx");
const { ModelPanel } = app("components/model-panel.tsx");
const { AdminPageFrame } = app("components/admin-chrome.tsx");
const { configurationPatch } = app("lib/admin-config.ts");

test("configuration validation is exact, typed and prototype-safe", () => {
  const patch = configurationPatch({ title: "string", count: "int", enabled: "bool", disabled: "bool" }, { title: "", count: "-42", enabled: "true", disabled: "false" });
  assert.equal(Object.getPrototypeOf(patch), null);
  assert.deepEqual({ ...patch }, { title: "", count: -42, enabled: true, disabled: false });
  for (const value of ["1.5", "1e2", "9007199254740992", ""]) assert.throws(() => configurationPatch({ n: "int" }, { n: value }), /integer/);
  assert.throws(() => configurationPatch({ b: "bool" }, { b: "yes" }), /true or false/);
  assert.throws(() => configurationPatch({ a: "secret" }, { a: "value" }), /Unsupported/);
  assert.throws(() => configurationPatch({}, { constructor: "value" }), /Unknown/);
});

test("result table pages, filters, and tabs respond to user events", async () => {
  const rows = [{ name: "Alpha", count: 1 }, { name: "Beta", count: 2 }];
  await mount(AdminResultRenderer, { value: rows, spec: { type: "table", page_size: 1 } });
  assert.match(text(), /Alpha/);
  assert.doesNotMatch(text(), /Beta/);
  await click(button("Next"));
  assert.match(text(), /Beta/);
  await click(button("Previous"));
  assert.match(text(), /Alpha/);
  await change(find("input"), "beta");
  assert.match(text(), /Beta/);
  assert.doesNotMatch(text(), /Alpha/);
  await change(find("input"), "missing");
  assert.match(text(), /No rows/);
  await mount(AdminResultRenderer, { value: { count: 2, name: "Alpha" }, spec: { type: "tabs", children: [{ type: "metric", title: "Count", path: "count" }, { type: "detail", title: "Name", path: "name" }] } });
  await click(button("Name"));
  assert.equal(find('[role="tabpanel"]').textContent, "NameAlpha");
});

test("configuration submits only edited typed fields after confirmation and preserves errors", async () => {
  const calls = [];
  let fail = false;
  stubFetch((url, init) => { calls.push({ url, body: JSON.parse(init.body), auth: init.headers.authorization }); return fail ? new Response("denied", { status: 403 }) : json({ status: "success" }); });
  const props = { token: "fixture-token", projectID: "dev", pluginID: "sample", version: "1", schema: { enabled: "bool", limit: "int", name: "string", unsupported: "secret" } };
  await mount(AdminConfigForm, props);
  assert.equal(button("Save configuration").disabled, true);
  await change(find("select"), "true");
  await change(find('input[type="number"]'), "7");
  await change(find('input[type="text"]:not([disabled])'), "name");
  await click(button("Keep unchanged"));
  window.confirm = () => false;
  await submit(find("form"));
  assert.equal(calls.length, 0);
  window.confirm = () => true;
  await submit(find("form"));
  assert.deepEqual(calls[0], { url: "/api/admin/actions", auth: "Bearer fixture-token", body: { operation: "plugin.configure", project_id: "dev", plugin_id: "sample", version: "1", values: { limit: 7, name: "name" }, confirmed: true } });
  assert.match(text(), /Configuration updated/);
  await change(find("select"), "false");
  fail = true;
  await submit(find("form"));
  assert.match(text(), /denied/);
  await change(find("select"), "");
  assert.equal(button("Save configuration").disabled, true);
});

const action = (operation, title, read_only, fields = []) => ({ operation, title, read_only, fields });
const contribution = (page, pluginID = "sample") => ({ projectID: "dev", pluginID, version: "1", page, href: "/extensions/dev/" + pluginID + "/1/" + page.id });

test("declarative admin action forms validate inputs, distinguish reads and mutations, and render results", async () => {
  session();
  window.confirm = () => true;
  const calls = [];
  const page = contribution({ id: "main", title: "Sample", actions: [
    action("sample.read", "Read", true, [{ name: "name", type: "string", required: true }, { name: "limit", type: "number", required: false }, { name: "enabled", type: "boolean", required: false }, { name: "data", type: "json", required: false }]),
    action("sample.write", "Write", false),
  ] });
  stubFetch((url, init) => { calls.push(JSON.parse(init.body)); return reply({ ok: true }); });
  await mount(AdminPageRenderer, { contribution: page });
  await submit(find("form"));
  assert.match(text(), /name is required/);
  await change(find('input[type="text"]'), "test");
  await change(find('input[type="number"]'), "3");
  const selects = [...document.querySelectorAll("select")];
  await change(selects[1], "false");
  await change(find("textarea"), "{");
  await submit(find("form"));
  assert.equal(calls.length, 0);
  await change(find("textarea"), '{"x":1}');
  await submit(find("form"));
  assert.equal(calls[0].operation, "plugin.admin.query");
  assert.deepEqual(calls[0].payload, { name: "test", limit: 3, enabled: false, data: { x: 1 } });
  assert.match(text(), /"ok": true/);
  await change(find("select"), "sample.write");
  window.confirm = () => false;
  await submit(find("form"));
  assert.equal(calls.length, 1);
  window.confirm = () => true;
  await change(find("textarea"), '{"value":2}');
  await submit(find("form"));
  assert.equal(calls[1].operation, "plugin.admin.invoke");
  assert.equal(calls[1].confirmed, true);
  assert.deepEqual(calls[1].payload, { value: 2 });
});

test("inventory selects rows, paginates, refreshes and shows a missing session", async () => {
  session();
  const selected = [], cursors = [];
  const page = contribution({ id: "files", title: "Files", actions: [], inventory: { title: "Objects", kind: "file_keys" } });
  stubFetch((url, init) => {
    const body = JSON.parse(init.body); cursors.push(body.after);
    return json(body.after ? { rows: [] } : { rows: [{ key: "a", size_bytes: 3, published: true }, { key: "b", published: false }, { key: "c" }], next: "c" });
  });
  await mount(AdminInventory, { contribution: page, refresh: 0, disabled: false, onSelect: key => selected.push(key) });
  assert.match(text(), /Objects/);
  assert.match(text(), /Yes/);
  assert.match(text(), /No/);
  await click(button("Select"));
  assert.deepEqual(selected, ["a"]);
  await click(button("Next"));
  assert.match(text(), /No resources/);
  await click(button("Previous"));
  await click(button("Refresh"));
  assert.deepEqual(cursors, ["", "c", "", ""]);
  window.sessionStorage.clear();
  await mount(AdminInventory, { contribution: page, refresh: 1, disabled: false });
  assert.match(text(), /Sign in to continue/);
  await mount(AdminInventory, { contribution: contribution({ id: "empty", title: "Empty", actions: [] }), refresh: 0, disabled: false });
  assert.equal(text(), "");
});

test("model form enforces revisions and sends typed create and update fields", async () => {
  session();
  window.confirm = () => true;
  let record = { id: "one", revision: "r1", state: "draft", fields: { count: 1, enabled: true, name: "old" } };
  const page = contribution({ id: "model", title: "Model", model: { name: "items", schema: { version: 1, initial: "draft", fields: { count: "integer", enabled: "boolean", name: "string" } } }, actions: [
    { ...action("sample.list", "List", true), model_action: "list" },
    { ...action("sample.create", "Create", false), model_action: "create" },
    { ...action("sample.update", "Update", false), model_action: "update" },
    { ...action("sample.delete", "Delete", false), model_action: "delete" },
  ] });
  const calls = [];
  stubFetch((url, init) => {
    const body = JSON.parse(init.body); calls.push(body);
    if (body.target_operation === "sample.list") return reply([record]);
    if (body.target_operation === "sample.delete") return reply(null);
    record = { ...record, id: body.payload.id, revision: "r2", fields: body.payload.fields };
    return reply(record);
  });
  await mount(ModelPanel, { contribution: page });
  assert.equal(button("Update").disabled, true);
  await click(button("Select"));
  assert.equal(button("Update").disabled, false);
  await change(find('input[type="number"]'), "2");
  await change(find("select"), "false");
  await click(button("Update"));
  assert.deepEqual(calls.at(-1).payload, { id: "one", revision: "r1", fields: { count: 2, enabled: false, name: "old" } });
  await click(button("Delete"));
  assert.deepEqual(calls.at(-1).payload, { id: "one", revision: "r2" });
  assert.equal(button("Update").disabled, true);
  await change(find('input[type="text"]'), "two");
  await change(find('input[type="number"]'), "5");
  await click(button("Create"));
  assert.equal(calls.at(-1).payload.id, "two");
});

test("protected admin frame checks live installation before mounting content", async () => {
  const props = { activePage: "functions", title: "Functions", requiredPluginID: "functions", children: "protected content" };
  await mount(AdminPageFrame, props);
  assert.deepEqual(navigations.at(-1), ["replace", "/"]);
  assert.doesNotMatch(text(), /protected content/);
  session();
  let enabled = false;
  stubFetch(url => url.includes("/projects") ? json([{ ID: "dev", Active: true }]) : json({ installations: [{ PluginID: "functions", Enabled: enabled, Ready: true }] }));
  await mount(AdminPageFrame, props);
  assert.doesNotMatch(text(), /protected content/);
  enabled = true;
  await mount(AdminPageFrame, props);
  assert.match(text(), /protected content/);
});
