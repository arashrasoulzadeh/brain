import { validateProjectId } from "../sdk-typescript/src/index.js";
import type { FunctionPublicationState } from "./function-publication-state.js";
import { FunctionPublicationReader } from "./function-publication-reader.js";
export class FunctionPublicationRollback {
  private readonly reader: FunctionPublicationReader;
  constructor(private readonly state: FunctionPublicationState) {
    this.reader = new FunctionPublicationReader(state);
  }
  rollback(projectId: string, functionId: string, version: number) {
    validateProjectId(projectId);
    if (!functionId || !Number.isSafeInteger(version) || version < 1)
      throw new Error("invalid function rollback");
    const publication = this.reader.get(projectId, functionId, version);
    if (!publication) throw new Error("function publication not found");
    this.state.activeVersions.set(`${projectId}:${functionId}`, version);
    return publication;
  }
}
