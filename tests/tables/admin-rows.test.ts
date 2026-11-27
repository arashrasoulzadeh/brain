import assert from "node:assert/strict";
import { test } from "node:test";
import { mutateTableRow, read } from "../../plugins/tables/index.js";
import type { PluginContext } from "../../plugins/sdk-typescript/src/index.js";

test("bounded table reads preserve host revision and pagination metadata", () => {
  const rows = [{ row_index: 1, row: { name: "test" } }];
  assert.deepEqual(read({ table: "messages", rows, page: { revision: "r", total_rows: 3, offset: 1, next_offset: 2 } }), { table: "messages", rows, revision: "r", total_rows: 3, offset: 1, next_offset: 2 });
  assert.deepEqual(read({ table: "messages", rows: [] }), { table: "messages", rows: [] });
});

test("row mutation worker checks the host snapshot and emits a refresh observation", () => {
  const events: string[] = [];
  const context: PluginContext = { project: { id: "alpha", active: true }, plugin: { id: "tables", version: "1" }, execution: { operation: "tables.update" }, plugins: { observe: { fire(name) { events.push(name); } } } };
  const result = { table: "messages", revision: "r", updated: true };
  assert.deepEqual(mutateTableRow(context, { projectId: "alpha", table: "messages", result }), result);
  assert.throws(() => mutateTableRow(context, { projectId: "beta", table: "messages", result }), /project context mismatch/);
  assert.throws(() => mutateTableRow(context, { table: "messages", result: { ...result, table: "other" } }), /snapshot required/);
  const deleting = { ...context, execution: { operation: "tables.delete" } };
  const deleted = { table: "messages", revision: "r2", deleted: true };
  assert.deepEqual(mutateTableRow(deleting, { table: "messages", result: deleted }), deleted);
  assert.deepEqual(events, ["record.written", "record.written"]);
});
