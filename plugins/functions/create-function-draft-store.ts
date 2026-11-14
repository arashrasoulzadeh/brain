import { createFunctionDraftState } from "./function-draft-state.js";
import { FunctionDraftWriter } from "./function-draft-writer.js";
import { FunctionDraftReader } from "./function-draft-reader.js";
export function createFunctionDraftStore() {
  const state = createFunctionDraftState();
  const writer = new FunctionDraftWriter(state);
  const reader = new FunctionDraftReader(state);
  return { save: writer.save.bind(writer), get: reader.get.bind(reader) };
}
