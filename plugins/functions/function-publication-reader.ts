import type { FunctionPublicationState } from "./function-publication-state.js";
import { validateProjectId } from "../sdk-typescript/src/index.js";
export class FunctionPublicationReader {
  constructor(private readonly state: FunctionPublicationState) {}
  get(projectId: string, functionId: string, version: number) {
    validateProjectId(projectId);
    const publication = this.state.publications.get(`${projectId}:${functionId}:${version}`);
    return publication ? { ...publication } : undefined;
  }
}
