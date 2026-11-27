import assert from "node:assert/strict";
import { test } from "node:test";
import {
  requireProjectInput,
  projectRoomID,
} from "../../plugins/sdk-typescript/src/validation.js";
import type { PluginContext } from "../../plugins/sdk-typescript/src/index.js";

test("project input requires an object and never trusts payload identity", () => {
  const context: PluginContext = {
    project: { id: "alpha", active: true },
    plugin: { id: "sample", version: "1" },
    execution: { operation: "read" },
    plugins: { observe: { fire() {} } },
  };
  for (const invalid of [null, 1, "alpha", [], ["alpha"]]) {
    assert.throws(() => requireProjectInput(context, invalid), /invalid plugin input/);
  }
  for (const key of ["projectId", "project_id"]) {
    assert.throws(
      () => requireProjectInput(context, { [key]: "beta" }),
      /host project context mismatch/,
    );
  }
  assert.deepEqual(requireProjectInput(context, { projectId: "alpha" }), { projectId: "alpha" });
  assert.equal(projectRoomID("alpha", "room"), "alpha:room");
  assert.throws(() => projectRoomID("bad/id", "room"), /invalid project id/);
  assert.throws(() => projectRoomID("alpha", "bad/id"), /invalid project room id/);
});
