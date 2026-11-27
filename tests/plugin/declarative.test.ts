import { test } from "node:test";
import assert from "node:assert/strict";
import { validateModel, transitionState } from "../../plugins/sdk-typescript/src/index.js";

test("model validation rejects extra fields and noninteger values", () => {
  const schema = { name: "string", count: "integer" } as const;
  assert.equal(validateModel(schema, { name: "room", count: 2 }).count, 2);
  assert.throws(() => validateModel(schema, { name: "room", count: 2.5 }));
  assert.throws(() => validateModel(schema, { name: "room", count: 2, projectId: "forged" }));
});

test("state transitions reject unknown actions and wrong source states", () => {
  const machine = { initial: "waiting", transitions: { start: { from: "waiting", to: "playing" } } };
  assert.equal(transitionState(machine, "waiting", "start"), "playing");
  assert.throws(() => transitionState(machine, "playing", "start"));
  assert.throws(() => transitionState(machine, "waiting", "constructor"));
});
