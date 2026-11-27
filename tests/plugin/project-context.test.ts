import { deepEqual, equal } from "node:assert/strict";
import { mkdtemp, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { spawnSync } from "node:child_process";
import { test } from "node:test";
import { createOperationRoutes } from "../../plugins/sdk-typescript/src/index.js";
import type { PluginContext } from "../../plugins/sdk-typescript/src/index.js";

test("worker injects project and plugin context across manifest execution modes", async () => {
  const directory = await mkdtemp(join(tmpdir(), "brain-project-context-"));
  try {
    const entrypoint = join(directory, "plugin.mjs");
    const descriptor = join(directory, "descriptor.json");
    await writeFile(entrypoint, [
      "export function dispatchRoute(name, input, context) { return context; }",
      "export function dispatchRequest(name, request, context) { return request.context === context ? context : null; }",
      "export function read(context) { return context; }",
      "export function handle(request, context) { return request.context === context ? context : null; }",
    ].join("\n"));
    const routes = [
      { operation: "dispatch", function: "read", helper: "sdk.dispatch" },
      { operation: "request", function: "read", helper: "sdk.request" },
      { operation: "context", function: "read", helper: "sdk.context" },
      { operation: "handle", function: "read" },
      { operation: "response", function: "read", helper: "sdk.response", response: { input: [], output: { project: "$project.id", plugin: "$plugin.id" } } },
    ];
    await writeFile(descriptor, JSON.stringify({ entrypoint, projectId: "trusted", pluginId: "sample", version: "1", requireProject: true, routes }));
    const params = { protocol: 1, projectId: "trusted", project: { id: "trusted", active: true }, pluginId: "sample", version: "1", payload: { projectId: "trusted" } };
    const requests = routes.map(({ operation }) => ({ id: operation, params: { ...params, operation }, context: { project: { id: "forged" } } }));
    requests.push({ id: "denied", params: { ...params, operation: "context", projectId: "other" }, context: { project: { id: "forged" } } });
    requests.push({ id: "forged-payload", params: { ...params, operation: "context", payload: { projectId: "forged" } }, context: { project: { id: "forged" } } });
    const missingProject = { id: "missing-model", params: { protocol: 1, projectId: "trusted", pluginId: "sample", version: "1", operation: "context" } };
    const worker = spawnSync(process.execPath, ["scripts/typescript-worker.mjs", descriptor], { input: [...requests, missingProject].map((request) => JSON.stringify(request)).join("\n") + "\n", encoding: "utf8", timeout: 5000 });
    equal(worker.status, 0, worker.stderr);
    const results = worker.stdout.trim().split("\n").map((line) => JSON.parse(line) as { result?: { payload: unknown }; errorCode?: string });
    for (const [index, operation] of ["dispatch", "request", "context", "handle"].entries()) {
      deepEqual(results[index].result?.payload, { project: { id: "trusted", active: true }, plugin: { id: "sample", version: "1" }, execution: { operation }, helpers: {}, plugins: { observe: {} }, auth: { actor: { kind: "public", id: "", authenticated: false, level: "anonymous" }, caller: { kind: "public", id: "", authenticated: false, level: "anonymous" }, permissions: [] } });
    }
    deepEqual(results[4].result?.payload, { project: "trusted", plugin: "sample" });
    equal(results[5].errorCode, "execution context mismatch");
    equal(results[6].errorCode, "host project context mismatch");
    equal(results[7].errorCode, "project model mismatch");
  } finally {
    await rm(directory, { recursive: true, force: true });
  }
});

test("shared operation dispatcher forwards context to business functions", () => {
  const context: PluginContext = { project: { id: "trusted", active: true }, plugin: { id: "sample", version: "1" }, execution: { operation: "read" }, plugins: { observe: { fire() {} } } };
  const routes = createOperationRoutes(["read"], (_operation, _input, executionContext) => executionContext);
  equal(routes.dispatch("read", {}, context), context);
});

test("direct exported routes receive host project and configuration context", async () => {
  const directory = await mkdtemp(join(tmpdir(), "brain-direct-context-"));
  try {
    const entrypoint = join(directory, "plugin.mjs");
    const descriptor = join(directory, "descriptor.json");
    await writeFile(entrypoint, "export function read(input, context) { return { project: context.project.id, active: context.project.active, limit: context.config.limit, dependency: context.plugins.has('other'), value: input.value }; }");
    await writeFile(descriptor, JSON.stringify({ entrypoint, projectId: "trusted", pluginId: "sample", version: "1", routes: [{ operation: "read", function: "read" }] }));
    const params = { protocol: 1, projectId: "trusted", pluginId: "sample", version: "1", operation: "read", payload: { projectId: "trusted", value: 7 }, hostContext: { projectVerified: true, invocationId: "host-one", config: { limit: 3 }, available: ["other@1"] } };
    const worker = spawnSync(process.execPath, ["scripts/typescript-worker.mjs", descriptor], { input: JSON.stringify({ id: "one", params }) + "\n", encoding: "utf8", timeout: 5000 });
    equal(worker.status, 0, worker.stderr);
    deepEqual(JSON.parse(worker.stdout).result.payload, { project: "trusted", active: true, limit: 3, dependency: true, value: 7 });
  } finally {
    await rm(directory, { recursive: true, force: true });
  }
});

test("worker uses only host actor and scoped permissions and never retains prior identity", async () => {
  const directory = await mkdtemp(join(tmpdir(), "brain-auth-context-"));
  try {
    const entrypoint = join(directory, "plugin.mjs");
    const descriptor = join(directory, "descriptor.json");
    await writeFile(entrypoint, "export function read(context) { const a = context.auth; return { actor: a.actor, caller: a.caller, allowed: a.can('read', 'one'), other: a.can('read', 'two'), frozen: Object.isFrozen(a) && Object.isFrozen(a.actor) && Object.isFrozen(a.permissions) }; }");
    await writeFile(descriptor, JSON.stringify({ entrypoint, projectId: "trusted", pluginId: "sample", version: "1", routes: [{ operation: "read", function: "read", helper: "sdk.context" }, { operation: "who", function: "read", helper: "sdk.response", response: { input: [], output: { id: "$auth.actor.id", caller: "$auth.caller.kind" } } }] }));
    const actor = { kind: "operator", id: "root", authenticated: true, level: "authenticated" };
    const caller = { kind: "plugin", id: "source@1", authenticated: false, level: "system" };
    const params = { protocol: 1, projectId: "trusted", pluginId: "sample", version: "1", operation: "read", payload: { auth: { actor } } };
    const auth = { actor, caller, permissions: [{ operation: "read", capability: "read", resource: "one" }] };
    const requests = [{ id: "1", params: { ...params, hostContext: { auth } } }, { id: "2", params, context: { auth } }, { id: "3", params: { ...params, operation: "who", hostContext: { auth } } }];
    const worker = spawnSync(process.execPath, ["scripts/typescript-worker.mjs", descriptor], { input: requests.map((request) => JSON.stringify(request)).join("\n") + "\n", encoding: "utf8", timeout: 5000 });
    equal(worker.status, 0, worker.stderr);
    const results = worker.stdout.trim().split("\n").map((line) => JSON.parse(line));
    deepEqual(results[0].result.payload, { actor, caller, allowed: true, other: false, frozen: true });
    equal(results[1].result.payload.actor.kind, "public");
    equal(results[1].result.payload.allowed, false);
    deepEqual(results[2].result.payload, { id: "root", caller: "plugin" });
  } finally {
    await rm(directory, { recursive: true, force: true });
  }
});
