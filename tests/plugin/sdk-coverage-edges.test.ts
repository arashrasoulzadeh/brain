import assert from "node:assert/strict";
import { test } from "node:test";
import { validateModel, transitionState } from "../../plugins/sdk-typescript/src/declarative.js";
import { runCappedRetries, runWithTimeout, commitWithGrant, createTransactionGate, createCommitRollback, createFixedWindowLimiter } from "../../plugins/sdk-typescript/src/execution.js";
import { request, pluginRequest, createResourceRequest } from "../../plugins/sdk-typescript/src/request.js";
import { createRouteDispatcher, createRouteHandler, createHostHandler, createFunctionHandler, createRequestDispatcher } from "../../plugins/sdk-typescript/src/routes.js";

test("declarative models reject arrays, inherited fields, oversized models and invalid field types", () => {
  for (const input of [null, 1, [], "value"]) {
    assert.throws(() => validateModel({}, input), /expected model object/);
  }
  assert.throws(() => validateModel({ name: "string" }, {}), /invalid model fields/);
  const tooLarge = Object.fromEntries(Array.from({ length: 65 }, (_, i) => ["field" + i, "string" as const]));
  assert.throws(() => validateModel(tooLarge, tooLarge), /invalid model fields/);
  for (const key of ["bad-key", "constructor", "prototype", "__proto__"]) {
    assert.throws(() => validateModel({ [key]: "string" }, { [key]: "value" }), /invalid model field/);
  }
  assert.throws(() => validateModel({ name: "string" }, { other: "value" }), /invalid model field/);
  for (const bad of [NaN, Infinity, 1.5, "1"]) {
    assert.throws(() => validateModel({ count: "integer" }, { count: bad }), /invalid model field type/);
  }
  assert.throws(() => validateModel({ name: "string" }, { name: 1 }), /invalid model field type/);
  assert.throws(() => validateModel({ active: "boolean" }, { active: "true" }), /invalid model field type/);
  const value = validateModel({ name: "string", count: "integer", active: "boolean" }, { name: "a", count: 1, active: false });
  assert.deepEqual({ ...value }, { name: "a", count: 1, active: false });
  assert.equal(Object.getPrototypeOf(value), null);
  assert.equal(Object.isFrozen(value), true);
  const machine = { initial: "start", transitions: { advance: { from: "start", to: "end" }, broken: { from: "start", to: "" } } };
  assert.equal(transitionState(machine, "start", "advance"), "end");
  assert.throws(() => transitionState(machine, "end", "advance"), /invalid transition/);
  assert.throws(() => transitionState(machine, "start", "broken"), /invalid transition/);
  assert.throws(() => transitionState(machine, "start", "constructor"), /unknown transition/);
});

test("execution helpers reject invalid limits and preserve failed transaction semantics", async () => {
  for (const cap of [0, 1.5]) {
    await assert.rejects(runCappedRetries(() => 1, cap), /invalid retry cap/);
    await assert.rejects(runWithTimeout(Promise.resolve(1), cap), /invalid invocation timeout/);
  }
  await assert.rejects(runCappedRetries(() => { throw "failure"; }, 2), /retry attempts exhausted/);
  await assert.rejects(runCappedRetries(() => { throw new Error("specific"); }, 1), /specific/);
  assert.equal(await runCappedRetries(() => 42, 1), 42);
  assert.equal(await runWithTimeout(Promise.resolve(42), 1000), 42);
  await assert.rejects(runWithTimeout(new Promise(() => {}), 1), /timed out/);
  assert.throws(() => commitWithGrant(() => true, () => 1, ""), /grant revoked/);
  assert.throws(() => commitWithGrant(() => false, () => 1, "write"), /grant revoked/);
  assert.equal(commitWithGrant(() => true, () => 1, "write"), 1);
  const gate = createTransactionGate("closed");
  assert.throws(() => gate.finish(() => { throw new Error("commit failure"); }), /commit failure/);
  assert.throws(() => gate.ensureOpen(), /closed/);
  let commits = 0;
  const transaction = createCommitRollback(() => { commits++; });
  transaction.commit();
  assert.throws(() => transaction.rollback(), /transaction already finished/);
  const rollback = createCommitRollback(() => { commits++; });
  rollback.rollback();
  assert.throws(() => rollback.commit(), /transaction already finished/);
  assert.equal(commits, 1);
  const admit = createFixedWindowLimiter(2, 100);
  assert.equal(admit("a", 0), true);
  assert.equal(admit("a", 1), true);
  assert.equal(admit("a", 2), false);
  assert.equal(admit("b", 2), true);
  assert.equal(admit("a", 100), true);
});

test("request serialization and dispatch reject incomplete metadata and undeclared routes", () => {
  const params = { protocol: 1 as const, projectId: "dev", pluginId: "sample", version: "1", operation: "read" };
  assert.throws(() => request("", params), /invalid plugin request/);
  for (const key of ["projectId", "pluginId", "version", "operation"] as const) {
    assert.throws(() => request("id", { ...params, [key]: "" }), /invalid plugin request/);
  }
  assert.equal(pluginRequest("p", "dev", "sample", "1", "read").id, "p-read-dev");
  assert.deepEqual(createResourceRequest("p", "id")("dev", "sample", "1", "read", "a").params.payload, { id: "a" });
  const rpc = request("id", params);
  const valid = [{ operation: "read", function: "read", capability: "read" }];
  const context = { projectId: "dev", invocationId: "one" };
  assert.deepEqual(createHostHandler({ invoke: (...args) => args }, valid, context)(rpc), {
    accepted: true, operation: "read", payload: ["dev", "read", "read", null],
  });
  for (const invalid of [{ projectId: "", invocationId: "one" }, { projectId: "dev", invocationId: "" }]) {
    assert.throws(() => createHostHandler({ invoke: () => null }, valid, invalid)(rpc), /host project context mismatch/);
  }
  for (const route of [{ ...valid[0], function: "" }, { ...valid[0], capability: "" }]) {
    assert.throws(() => createRouteHandler([route])(rpc), /operation not declared/);
    assert.throws(() => createFunctionHandler({}, [route], context)(rpc), /invalid plugin route/);
  }
  assert.throws(() => createRouteDispatcher({ read: undefined as never })("read", null), /not exported/);
  assert.throws(() => createRequestDispatcher({})("read", rpc), /not exported/);
});
