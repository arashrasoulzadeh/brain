import { createProjectMap, validateProjectId } from "../sdk-typescript/src/index.js";
import type { FunctionDraft, FunctionPublication } from "./index.js";

export function createFunctionPublicationStore() {
  const publications = createProjectMap<Map<number, FunctionPublication>>();
  const activeVersions = createProjectMap<number>();

  function get(projectId: string, functionId: string, version: number) {
    const publication = publications.get(projectId, functionId)?.get(version);
    return publication ? { ...publication } : undefined;
  }

  return {
    publish(
      draft: FunctionDraft,
      artifactDigest: string,
      toolchain: string,
      publishedAt = Date.now(),
    ): FunctionPublication {
      validateProjectId(draft.projectId);
      if (
        !draft.functionId ||
        !draft.source ||
        !/^sha256:[a-f0-9]+$/.test(artifactDigest) ||
        !/^go[0-9]+\.[0-9]+(?:\.[0-9]+)?$/.test(toolchain) ||
        !Number.isFinite(publishedAt)
      )
        throw new Error("invalid function publication");
      const versions = publications.get(draft.projectId, draft.functionId) ?? new Map<number, FunctionPublication>();
      const version = versions.size ? Math.max(...versions.keys()) + 1 : 1;
      const publication = { ...draft, version, artifactDigest, toolchain, publishedAt };
      versions.set(version, publication);
      publications.set(draft.projectId, draft.functionId, versions);
      activeVersions.set(draft.projectId, draft.functionId, version);
      return { ...publication };
    },
    get,
    active(projectId: string, functionId: string) {
      const version = activeVersions.get(projectId, functionId);
      return version === undefined ? undefined : get(projectId, functionId, version);
    },
    rollback(projectId: string, functionId: string, version: number) {
      validateProjectId(projectId);
      if (!functionId || !Number.isSafeInteger(version) || version < 1)
        throw new Error("invalid function rollback");
      const publication = get(projectId, functionId, version);
      if (!publication) throw new Error("function publication not found");
      activeVersions.set(projectId, functionId, version);
      return publication;
    },
  };
}
