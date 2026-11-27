import assert from "node:assert/strict";
import { test } from "node:test";
import { app, mount, find, button, click, change, text, session, stubFetch, json, submit, navigations } from "./dom-harness.mjs";

const Detail = app("functions/[name]/view.tsx").default;
const Create = app("functions/new/view.tsx").default;
const Executions = app("functions/function-executions.tsx").FunctionExecutions;
const draft = { project_id: "dev", name: "echo", version: 1, method: "POST", public: false, permissions: ["functions.invoke"], source: "export const config = { method: 'POST' }; export default function handler(input) { return input; }", created_at: "2026-01-01T00:00:00Z" };
const catalog = { manifests: [{ ID: "functions", Version: "1", Capabilities: [] }], installations: [{ ProjectID: "dev", PluginID: "functions", Version: "1", Enabled: true, Ready: true }] };

test("function workspace executes current source, rejects malformed JSON, and confirms publishing", async () => {
  session();
  const calls = [];
  stubFetch((url, init) => {
    if (url === "/_system/admin/projects") return json([{ ID: "dev", Slug: "alias", Active: true }]);
    if (url === "/dev/admin/plugins") return json(catalog);
    if (url === "/dev/admin/functions" && init.method !== "POST") return json([draft, { ...draft, project_id: "foreign" }]);
    if (url === "/dev/admin/functions") {
      const body = JSON.parse(init.body);
      calls.push(body);
      if (body.operation === "functions.run") return json({ valid: true, project_id: "dev", name: "echo", result: { echo: body.input } });
      if (body.operation === "functions.test") return json({ valid: true, project_id: "dev", name: "echo" });
      if (body.operation === "functions.publish") return json({ published: true, project_id: "dev", name: "echo", version: 1, method: "POST", published_at: "2026-01-02T00:00:00Z" });
      if (body.operation === "functions.permissions") return new Response(null, { status: 204 });
    }
    throw new Error("unexpected request " + url);
  });
  await mount(Detail, {}, { pathname: "/alias/ui/functions/echo", params: { name: "echo" } });
  assert.match(text(), /Test console/);
  await change(find(".fn-test textarea"), "{");
  await click(button("Run test"));
  assert.match(text(), /valid JSON/);
  assert.equal(calls.length, 0);
  await change(find(".fn-test textarea"), '{"value":42}');
  await click(button("Run test"));
  assert.match(text(), /Execution succeeded/);
  assert.match(text(), /42/);
  assert.equal(calls[0].operation, "functions.run");
  assert.equal(calls[0].source, draft.source);
  assert.equal(calls[0].project_id, "dev");
  assert.deepEqual(calls[0].input, { value: 42 });
  await click(button("Check syntax"));
  assert.match(text(), /Host preflight passed/);
  await click(button("Publish current version"));
  assert.equal(calls.length, 2);
  await click(button("Confirm publish"));
  assert.equal(calls.at(-1).operation, "functions.publish");
  assert.match(text(), /Publication confirmed/);
  await click(button("Publication"));
  assert.equal(find(".fn-panel input[readonly]").value, "http://localhost:3000/alias/functions/echo");
  await click(button("Settings"));
  await change(find(".fn-panel input:not([type=checkbox])"), "functions.invoke, extra");
  await click(button("Save settings"));
  assert.equal(calls.at(-1).operation, "functions.publish");
  await click(button("Confirm settings"));
  assert.equal(calls.at(-1).operation, "functions.permissions");
  assert.deepEqual(calls.at(-1).permissions, ["extra", "functions.invoke"]);
});

test("creation uses canonical scope and starter source", async () => {
  session();
  const calls = [];
  stubFetch((url, init) => {
    if (url === "/_system/admin/projects") return json([{ ID: "dev", Slug: "alias", Active: true }]);
    if (url === "/dev/admin/plugins") return json(catalog);
    if (url === "/dev/admin/functions") {
      const body = JSON.parse(init.body);
      calls.push(body);
      return json({ ...draft, name: body.name, source: body.source, version: body.function_version, method: "GET" });
    }
    throw new Error("unexpected request " + url);
  });
  await mount(Create, {}, { pathname: "/alias/ui/functions/new" });
  assert.equal(button("Create function").disabled, true);
  await click(find(".fn-template:nth-child(2)"));
  await change(find('input[placeholder="hello-world"]'), "health");
  await submit(find("form.fn-studio-grid"));
  assert.equal(calls.length, 1);
  assert.equal(calls[0].project_id, "dev");
  assert.equal(calls[0].operation, "functions.create");
  assert.match(calls[0].source, /method: 'GET'/);
  assert.equal(calls[0].public, false);
  assert.deepEqual(navigations.at(-1), ["push", "/dev/ui/functions/health"]);
});

test("execution records enforce project/function scope and expose step details", async () => {
  session();
  const logs = [
    { project_id: "dev", name: "echo", status: "success", duration_ms: 10, created_at: "2026-01-01", request_id: "first", steps: [{ name: "handler execution", status: "success", duration_ms: 5 }] },
    { project_id: "dev", name: "echo", status: "failed", duration_ms: 20, created_at: "2026-01-02", request_id: "failed", error: "bad input" },
    { project_id: "foreign", name: "echo", status: "success", duration_ms: 1, created_at: "2026-01-03", request_id: "foreign-secret" },
    { project_id: "dev", name: "other", status: "success", duration_ms: 1, created_at: "2026-01-03", request_id: "other-secret" },
  ];
  stubFetch(url => { assert.equal(url, "/dev/admin/function-logs?name=echo"); return json(logs); });
  await mount(Executions, { scope: { project: "dev", namespace: "alias", token: "fixture-token" }, name: "echo", revision: 0, analytics: false });
  assert.match(text(), /50%/);
  assert.doesNotMatch(text(), /foreign-secret|other-secret/);
  await change(find('input[type="search"]'), "first");
  await click(button("View"));
  assert.match(text(), /handler execution/);
  await click(find('button[aria-label="Close dialog"]'));
  await change(find('input[type="search"]'), "");
  await change(find("select"), "failure");
  assert.match(text(), /bad input/);
  assert.doesNotMatch(text(), /first/);
});
