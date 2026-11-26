import assert from "node:assert/strict";
import { test } from "node:test";
import { configurationPatch } from "../../admin-nextjs/app/lib/admin-config.js";

test("configuration patches preserve types and reject undeclared fields", () => {
  const schema = { title: "string", count: "int", enabled: "bool" };
  assert.deepEqual({ ...configurationPatch(schema, { title: "", count: "4", enabled: "false" }) }, { title: "", count: 4, enabled: false });
  assert.throws(() => configurationPatch(schema, { count: "2.5" }));
  assert.throws(() => configurationPatch(schema, { enabled: "yes" }));
  assert.throws(() => configurationPatch(schema, { unknown: "x" }));
});
