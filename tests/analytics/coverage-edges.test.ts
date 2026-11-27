import assert from "node:assert/strict";
import { test } from "node:test";
import { AnalyticsRouteSink, analyticsRouteSink, capture, captureForProject, observeRoomConnected, readForProject, validateEvent } from "../../plugins/analytics/index.js";
import type { AnalyticsEvent } from "../../plugins/analytics/index.js";
import type { PluginContext } from "../../plugins/sdk-typescript/src/index.js";

test("analytics validates field limits, types and project-bound observations", () => {
  const valid = { projectId: "alpha", name: "test.event", fields: { count: 0, flag: false, text: "" } };
  validateEvent(valid, ["count", "flag", "text"]);
  for (const fields of [
    Object.fromEntries(Array.from({ length: 21 }, (_, i) => ["field" + i, i])),
    { invalid: undefined }, { invalid: null }, { invalid: {} }, { invalid: Infinity },
  ]) {
    assert.throws(() => validateEvent({ ...valid, fields } as AnalyticsEvent), /invalid event field/);
  }
  const sink = new AnalyticsRouteSink();
  capture(valid, sink);
  valid.fields.count = 10;
  assert.equal(sink.events[0].fields.count, 0);
  const event = { ...valid, fields: { count: 3 } };
  sink.append(event);
  event.fields.count = 4;
  assert.equal(sink.events[1].fields.count, 3);

  const observed: string[] = [];
  const context: PluginContext = {
    project: { id: "alpha", active: true },
    plugin: { id: "analytics", version: "1" },
    execution: { operation: "capture" },
    plugins: { observe: { fire(name) { observed.push(name); } } },
  };
  for (const input of [{ name: 7, fields: {} }, { name: "event" }, { name: "event", fields: "wrong" }]) {
    assert.throws(() => captureForProject(context, input), /invalid analytics event/);
  }
  assert.throws(() => captureForProject(context, { ...valid, projectId: "beta" }), /project context mismatch/);
  assert.deepEqual(captureForProject(context, { name: "event", fields: {} }), { captured: true, name: "event" });
  assert.deepEqual(captureForProject(context, valid), { captured: true, name: valid.name });
  const observation = { id: "one", projectId: "alpha", pluginId: "rooms", version: "1", event: "connected", payload: { roomId: "room", reconnected: false } };
  for (const input of [{ ...observation, projectId: "beta" }, { ...observation, pluginId: "other" }, { ...observation, event: "created" }, { ...observation, projectId: undefined }]) {
    assert.throws(() => observeRoomConnected(context, input), /project context mismatch|invalid room observation/);
  }
  assert.deepEqual(observeRoomConnected(context, observation), { captured: true, name: "rooms.connected" });
  assert.equal(analyticsRouteSink.events.at(-1)?.projectId, "alpha");
  assert.deepEqual(observed, ["captured", "captured"]);
});

test("analytics reads the host-provided project snapshot", () => {
  const context: PluginContext = { project: { id: "alpha" }, plugin: { id: "analytics", version: "1" }, execution: { operation: "analytics.history" }, plugins: { observe: { fire() {} } } };
  const result = { total_events: 2, rows: [] };
  assert.deepEqual(readForProject(context, { projectId: "alpha", result }), result);
  assert.throws(() => readForProject(context, { projectId: "beta", result }), /project context mismatch/);
  assert.throws(() => readForProject(context, { projectId: "alpha" }), /snapshot unavailable/);
  assert.throws(() => readForProject(context, { projectId: "alpha", result: [] }), /snapshot unavailable/);
});
