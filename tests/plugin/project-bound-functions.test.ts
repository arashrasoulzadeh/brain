import { test } from "node:test";
import assert from "node:assert/strict";
import { bindProjectSchedule, bindProjectTriggers, bindProjectJobLeases, createScheduleStore, createRecordChangeTriggerStore, createJobLeaseStore, createInvocationCommitStore, runProjectWorkflow } from "../../plugins/functions/index.js";
import { captureForProject, observeRoomConnected } from "../../plugins/analytics/index.js";
import { observeRoomConnected as templateObserver } from "../../plugins/template/index.js";
import type { PluginContext } from "../../plugins/sdk-typescript/src/index.js";

function context(id: string, active = true, events: string[] = []): PluginContext {
  return { project: { id, active }, plugin: { id: "functions", version: "1" }, execution: { operation: "test" }, plugins: { observe: { fire: event => { events.push(event); } } } };
}

test("scheduled and workflow helper bindings isolate projects", async () => {
  const a = context("alpha");
  const b = context("beta");
  const schedules = createScheduleStore();
  const first = bindProjectSchedule(a, schedules);
  const second = bindProjectSchedule(b, schedules);
  first.add("same", "* * * * *", 1);
  second.add("same", "* * * * *", 2);
  assert.deepEqual(first.due(3).map(s => s.projectId), ["alpha"]);
  assert.deepEqual(second.due(3).map(s => s.projectId), ["beta"]);
  assert.equal(first.claim("same", "worker", 3, 10), true);
  assert.equal(second.claim("same", "worker", 3, 10), true);
  assert.throws(() => bindProjectSchedule(context("alpha", false), schedules));
  const triggers = createRecordChangeTriggerStore();
  bindProjectTriggers(a, triggers).add("records", "write", { functionId: "job", version: 1 });
  assert.equal(bindProjectTriggers(b, triggers).match("records", "write").length, 0);
  const leases = createJobLeaseStore();
  assert.equal(bindProjectJobLeases(a, leases).acquire("job", "worker", 3, 10), true);
  assert.equal(bindProjectJobLeases(b, leases).acquire("job", "worker", 3, 10), true);
  const commits = createInvocationCommitStore();
  assert.equal(await runProjectWorkflow(a, commits, "same", ctx => ctx.project.id, 1), "alpha");
  assert.equal(await runProjectWorkflow(b, commits, "same", ctx => ctx.project.id, 1), "beta");
});

test("analytics direct capture publishes but observation capture never loops", () => {
  const events: string[] = [];
  const ctx = context("alpha", true, events);
  captureForProject(ctx, { name: "test.capture", fields: {} });
  const observation = { id: "event", projectId: "alpha", pluginId: "rooms", version: "1", event: "connected", payload: { roomId: "room", reconnected: false } };
  observeRoomConnected(ctx, observation);
  assert.deepEqual(events, ["captured"]);
  assert.deepEqual(templateObserver(ctx, observation), { recorded: true });
  assert.throws(() => templateObserver(context("beta"), observation));
});
