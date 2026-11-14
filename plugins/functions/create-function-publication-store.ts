import { createFunctionPublicationState } from "./function-publication-state.js";
import { FunctionPublicationWriter } from "./function-publication-writer.js";
import { FunctionPublicationReader } from "./function-publication-reader.js";
import { FunctionPublicationRollback } from "./function-publication-rollback.js";
export function createFunctionPublicationStore() {
  const state = createFunctionPublicationState();
  const writer = new FunctionPublicationWriter(state);
  const reader = new FunctionPublicationReader(state);
  const rollback = new FunctionPublicationRollback(state);
  return {
    publish: writer.publish.bind(writer),
    get: reader.get.bind(reader),
    active: (projectId: string, functionId: string) => {
      const version = state.activeVersions.get(`${projectId}:${functionId}`);
      return version === undefined ? undefined : reader.get(projectId, functionId, version);
    },
    rollback: rollback.rollback.bind(rollback),
  };
}
