import assert from "node:assert/strict";
import { test } from "node:test";
import { createProjectRequestGate } from "../../admin-nextjs/app/lib/project-requests.js";

test("late project responses cannot overwrite the selected project's catalog or functions", () => {
  const gate = createProjectRequestGate();
  assert.equal(gate.current(), "");
  assert.equal(gate.select("alpha"), true);
  const oldCatalog = gate.begin("alpha", "catalog")!;
  const oldFunctions = gate.begin("alpha", "functions")!;
  assert.ok(oldCatalog() && oldFunctions());
  assert.equal(gate.select("beta"), true);
  const latest = gate.begin("beta", "catalog")!;
  assert.equal(gate.begin("alpha", "catalog"), undefined);
  assert.equal(oldCatalog(), false);
  assert.equal(oldFunctions(), false);
  assert.equal(latest(), true);
  assert.equal(gate.select("beta"), false);
  assert.equal(latest(), true);
  const refresh = gate.begin("beta", "catalog")!;
  assert.equal(latest(), false);
  assert.equal(refresh(), true);
  assert.equal(gate.select("alpha"), true);
  assert.equal(oldCatalog(), false, "returning to a project must not revive old requests");
  assert.equal(refresh(), false);
  assert.equal(gate.current(), "alpha");
});
