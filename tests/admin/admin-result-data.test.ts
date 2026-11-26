import assert from "node:assert/strict";
import { test } from "node:test";
import { readAdminResultPath, recordRows, displayAdminValue } from "../../admin-nextjs/app/lib/admin-result-data.js";

test("result paths are declarative own-property lookups", () => {
  const value = { data: { rows: [{ id: 1 }] } };
  assert.deepEqual(readAdminResultPath(value, "data.rows"), [{ id: 1 }]);
  assert.equal(readAdminResultPath(value, "constructor"), undefined);
  assert.equal(readAdminResultPath(value, "__proto__"), undefined);
  assert.equal(readAdminResultPath(value, "data.missing"), undefined);
  assert.equal(recordRows([1]), undefined);
  assert.deepEqual(recordRows([]), []);
  assert.equal(displayAdminValue(false), "false");
});
