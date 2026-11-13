import type { FunctionPublicationState } from "./function-publication-state.js";
export class FunctionPublicationReader { constructor(private readonly state: FunctionPublicationState) {} get(projectId: string, functionId: string, version: number) { const publication = this.state.publications.get(`${projectId}:${functionId}:${version}`); return publication ? { ...publication } : undefined; } }
