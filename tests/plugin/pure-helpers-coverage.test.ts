import assert from "node:assert/strict";
import { test } from "node:test";
import { callHelper, hasHelper, helperNames } from "../../plugins/sdk-typescript/src/helpers.mjs";
import type { PureHelperName } from "../../plugins/sdk-typescript/src/helpers.mjs";

test("pure helper registry rejects invalid types and unknown helpers", () => {
  assert.equal(hasHelper("string.trim"), true);
  assert.equal(hasHelper("missing"), false);
  assert.equal(Object.isFrozen(helperNames()), true);
  for (const name of ["string.trim", "string.nonempty", "string.lower", "string.upper", "number.integer", "boolean.require"] as const) {
    assert.throws(() => callHelper(name, {}), /expected/);
  }
  assert.throws(() => callHelper("string.nonempty", " "), /expected nonempty string/);
  assert.throws(() => callHelper("number.integer", 1.5), /expected integer/);
  assert.throws(() => callHelper("unknown" as PureHelperName, null), /unknown pure helper/);
  assert.equal(callHelper("string.trim", " a "), "a");
  assert.equal(callHelper("string.nonempty", " a "), " a ");
  assert.equal(callHelper("string.lower", "A"), "a");
  assert.equal(callHelper("string.upper", "a"), "A");
  assert.equal(callHelper("number.integer", 0), 0);
  assert.equal(callHelper("boolean.require", false), false);
  assert.equal(callHelper("value.identity", null), null);
});
