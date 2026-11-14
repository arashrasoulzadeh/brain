import { createSQLChangeState } from "./sql-change-state.js";
import { SQLChangeWriter } from "./sql-change-writer.js";
import { SQLChangeReader } from "./sql-change-reader.js";
export function createDirectSQLChangeCapture() {
  const state = createSQLChangeState();
  const writer = new SQLChangeWriter(state);
  const reader = new SQLChangeReader(state);
  return { begin: writer.begin.bind(writer), list: reader.list.bind(reader) };
}
