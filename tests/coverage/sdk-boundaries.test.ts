import assert from "node:assert/strict";
import { test } from "node:test";
import {
  callHelper, hasHelper, helperNames, validateModel, transitionState, request, createFunctionHandler,
  requireProjectInput, runCappedRetries, validateBoundedInteger, type PluginContext,
} from "../../plugins/sdk-typescript/src/index.js";
import * as analytics from "../../plugins/analytics/index.js";
import * as files from "../../plugins/files/routes.js";
import * as functions from "../../plugins/functions/routes.js";
import * as rooms from "../../plugins/rooms/routes.js";
import * as matchmaking from "../../plugins/matchmaking/routes.js";
import * as tables from "../../plugins/tables/routes.js";
import { createVersionedRecord } from "../../plugins/tables/records.js";
import { planSchemaChange } from "../../plugins/tables/schema.js";

function context(id: string, events: unknown[]): PluginContext {
  return {
    project: { id: "coverage", active: true },
    plugin: { id, version: "1" },
    execution: { operation: id + ".test" },
    plugins: { observe: { fire(event, payload) { events.push({ event, payload }); } } },
  };
}

test("pure helper registry rejects values of the wrong type and is immutable", () => {
  assert.deepEqual(helperNames(), ["boolean.require", "number.integer", "string.lower", "string.nonempty", "string.trim", "string.upper", "value.identity"]);
  assert.ok(Object.isFrozen(helperNames()));
  assert.equal(hasHelper("string.trim"), true);
  assert.equal(hasHelper("missing"), false);
  assert.equal(callHelper("string.trim", "  a  "), "a");
  assert.equal(callHelper("string.nonempty", " a "), " a ");
  assert.equal(callHelper("string.upper", "ab"), "AB");
  assert.equal(callHelper("string.lower", "AB"), "ab");
  assert.equal(callHelper("boolean.require", false), false);
  assert.equal(callHelper("number.integer", 0), 0);
  const object = {};
  assert.equal(callHelper("value.identity", object), object);
  for (const helper of ["string.trim", "string.upper", "string.lower", "string.nonempty"] as const) {
    assert.throws(() => callHelper(helper, false), /expected/);
  }
  assert.throws(() => callHelper("string.nonempty", " "), /nonempty/);
  assert.throws(() => callHelper("number.integer", 1.5), /integer/);
  assert.throws(() => callHelper("boolean.require", "true"), /boolean/);
  assert.throws(() => callHelper("missing" as "string.trim", ""), /unknown pure helper/);
});

test("SDK validates exact models, transitions, request identity and retry errors", async () => {
  const validated = validateModel({ name: "string", count: "integer", enabled: "boolean" }, { name: "", count: 0, enabled: false });
  assert.equal(Object.getPrototypeOf(validated), null);
  assert.ok(Object.isFrozen(validated));
  assert.deepEqual({ ...validated }, { name: "", count: 0, enabled: false });
  for (const value of [null, [], 42]) assert.throws(() => validateModel({}, value), /object/);
  assert.throws(() => validateModel({ id: "string" }, { other: "x" }), /field/);
  assert.throws(() => validateModel({ constructor: "string" }, { constructor: "x" }), /field/);
  assert.throws(() => validateModel({ "bad-key": "string" }, { "bad-key": "x" }), /field/);
  assert.throws(() => validateModel({ n: "integer" }, { n: 1.2 }), /type/);
  assert.throws(() => validateModel({ n: "integer" }, { n: "1" }), /type/);
  assert.throws(() => validateModel({ b: "boolean" }, { b: "false" }), /type/);
  assert.throws(() => validateModel(Object.fromEntries<"string">(Array.from({ length: 65 }, (_, i) => ["x" + i, "string"] as const)), {}), /fields/);
  const machine = { initial: "draft", transitions: { publish: { from: "draft", to: "published" }, invalid: { from: "draft", to: "" } } };
  assert.equal(transitionState(machine, "draft", "publish"), "published");
  assert.throws(() => transitionState(machine, "published", "publish"), /invalid/);
  assert.throws(() => transitionState(machine, "draft", "invalid"), /invalid/);
  assert.throws(() => transitionState(machine, "draft", "unknown"), /unknown/);
  const params = { protocol: 1, projectId: "coverage", pluginId: "sample", version: "1", operation: "read" };
  assert.equal(request("id", params).params.projectId, "coverage");
  assert.throws(() => request("", params), /invalid/);
  for (const key of ["projectId", "pluginId", "version", "operation"] as const) assert.throws(() => request("id", { ...params, [key]: "" }), /invalid/);
  const rpc = request("id", { ...params, payload: { count: 1 } });
  const handler = createFunctionHandler({ read: (value) => value.params.payload }, [{ operation: "read", function: "read", capability: "read" }], { projectId: "coverage", invocationId: "invocation" });
  assert.deepEqual(await handler(rpc), { accepted: true, operation: "read", payload: { count: 1 } });
  const ctx = context("sample", []);
  assert.throws(() => requireProjectInput(ctx, []), /invalid plugin input/);
  assert.throws(() => requireProjectInput(ctx, { project_id: "other" }), /mismatch/);
  assert.throws(() => requireProjectInput({ ...ctx, project: { id: "coverage", active: false } }, {}), /context required/);
  for (const [value, max] of [[-1, 2], [1, 0], [1, 1.5], [1.5, 2]]) assert.throws(() => validateBoundedInteger(value, max, "bounded"), /bounded/);
  assert.throws(() => validateBoundedInteger(0, 0, "bounded", 1), /bounded/);
  await assert.rejects(runCappedRetries(() => { throw "fixture"; }, 2), /retry attempts exhausted/);
});

test("analytics validates field shapes and records only project-bound observations", () => {
  const events: unknown[] = [];
  const ctx = context("analytics", events);
  const sink = new analytics.AnalyticsRouteSink();
  const valid = { projectId: "coverage", name: "room.opened", fields: { count: 1, live: true, label: "room" } };
  assert.deepEqual(analytics.capture(valid, sink), { captured: true, name: "room.opened" });
  assert.equal(sink.events.length, 1);
  assert.notEqual(sink.events[0].fields, valid.fields);
  assert.throws(() => analytics.validateEvent(valid, ["count"]), /unapproved/);
  assert.throws(() => analytics.validateEvent({ ...valid, fields: { invalid: undefined } } as unknown as analytics.AnalyticsEvent), /value/);
  assert.throws(() => analytics.validateEvent({ ...valid, fields: { invalid: Infinity } }), /value/);
  assert.throws(() => analytics.validateEvent({ ...valid, fields: { invalid: "x".repeat(1025) } }), /value/);
  assert.throws(() => analytics.validateEvent({ ...valid, fields: { invalid: {} } } as unknown as analytics.AnalyticsEvent), /value/);
  assert.throws(() => analytics.captureForProject(ctx, { name: 7, fields: {} }), /invalid/);
  assert.throws(() => analytics.captureForProject(ctx, { name: "event" }), /invalid/);
  assert.throws(() => analytics.captureForProject(ctx, { name: "event", fields: "wrong" }), /invalid/);
  assert.deepEqual(analytics.captureForProject(ctx, { name: "event", fields: {} }), { captured: true, name: "event" });
  assert.deepEqual(events, [{ event: "captured", payload: { name: "event" } }]);
  assert.throws(() => analytics.observeRoomConnected(ctx, { projectId: "coverage", pluginId: "other", event: "connected" }), /invalid room/);
  assert.throws(() => analytics.observeRoomConnected(ctx, { projectId: "coverage", pluginId: "rooms", event: "other" }), /invalid room/);
});

test("route dispatchers handle stored and host-prepared file, room, ticket and table payloads", () => {
  const events: unknown[] = [];
  const file = context("files", events);
  assert.throws(() => files.upload({ key: "coverage-file" }, file), /data/);
  assert.deepEqual(files.upload({ key: "coverage-file", data: "hello" }, file), { key: "coverage-file", uploaded: true, published: false });
  assert.deepEqual(files.download({ key: "coverage-file" }, file), { key: "coverage-file", data: "hello", published: false });
  assert.deepEqual(files.metadata({ key: "coverage-file", data: "abc", published: true }, file), { key: "coverage-file", size: 3, published: true });
  assert.deepEqual(files.download({ key: "coverage-file", data: "host", published: false }, file), { key: "coverage-file", data: "host", published: false });
  assert.deepEqual(files.publish({ key: "coverage-file" }, file), { key: "coverage-file", published: true });
  assert.deepEqual(files.publish({ key: "external-file", published: true }, file), { key: "external-file", published: true });
  assert.throws(() => files.download({ key: "missing-coverage-file" }), /not found/);
  const room = context("rooms", events);
  rooms.create({ id: "coverage-room" }, room);
  rooms.join({ id: "coverage-room", member: "a" }, room);
  assert.deepEqual(rooms.reconnect({ id: "coverage-room", member: "a", room: { id: "coverage-room", members: ["a"] } }, room), { id: "coverage-room", members: ["a"], member: "a", reconnected: true });
  assert.deepEqual(rooms.result({ id: "coverage-room", result: { score: 2 } }, room), { id: "coverage-room", members: ["a"], result: { score: 2 } });
  assert.deepEqual(rooms.result({ id: "coverage-room" }, room), { id: "coverage-room", members: ["a"], result: { score: 2 } });
  const ticket = context("matchmaking", events);
  matchmaking.queue({ id: "coverage-ticket" }, ticket);
  assert.deepEqual(matchmaking.status({ id: "coverage-ticket" }, ticket), { id: "coverage-ticket", status: "queued" });
  assert.deepEqual(matchmaking.assign({ id: "coverage-ticket", ticket: { id: "coverage-ticket", status: "assigned" } }, ticket), { id: "coverage-ticket", status: "assigned" });
  assert.deepEqual(matchmaking.cancel({ id: "coverage-ticket" }, ticket), { id: "coverage-ticket", status: "cancelled" });
  const table = context("tables", events);
  tables.write({ table: "coverage-table", row: { b: 2, a: 1 } }, table);
  tables.write({ table: "coverage-table", row: { a: 3 } }, table);
  assert.deepEqual(tables.schema({ table: "coverage-table", rows: [{ a: 1 }, { b: 2 }, null, [], "bad"] }), { table: "coverage-table", columns: ["a", "b"] });
  assert.deepEqual(tables.read({ table: "coverage-table", rows: [{ host: true }] }), { table: "coverage-table", rows: [{ host: true }] });
  assert.deepEqual(tables.schema({ table: "coverage-table" }), { table: "coverage-table", columns: ["b", "a"] });
  assert.ok(events.some(event => (event as { event: string }).event === "record.written"));
  assert.ok(events.some(event => (event as { event: string }).event === "result.committed"));
});

test("function route accepts only public call input and table revision checks are observable", () => {
  const events: unknown[] = [];
  const ctx = context("functions", events);
  assert.deepEqual(functions.call({ name: "echo" }, ctx), { name: "echo", version: "latest", input: null, accepted: true, route: "public" });
  assert.deepEqual(functions.call({ name: "echo", version: " 2 ", permission: "public", input: { a: 1 } }), { name: "echo", version: "2", input: { a: 1 }, accepted: true, route: "public" });
  assert.throws(() => functions.call({ name: "echo", permission: "private" }, ctx), /not public/);
  assert.throws(() => functions.call(null), /name/);
  assert.deepEqual(events, [{ event: "callAccepted", payload: {} }]);
  const record = createVersionedRecord({ count: 1 });
  assert.equal(record.version, 0);
  record.update(0, { count: 2 });
  assert.equal(record.version, 1);
  assert.throws(() => record.update(0, { count: 3 }), /stale/);
  assert.throws(() => planSchemaChange("coverage", "table", { kind: "unsupported" } as unknown as Parameters<typeof planSchemaChange>[2]), /unsupported/);
});
