import { createApiChangeState } from "./api-change-state.js";
import { ApiChangeWriter } from "./api-change-writer.js";
import { ApiChangeReader } from "./api-change-reader.js";
export function createApiChangeCapture() {
  const state = createApiChangeState();
  const writer = new ApiChangeWriter(state);
  const reader = new ApiChangeReader(state);
  return { begin: writer.begin.bind(writer), list: reader.list.bind(reader) };
}
