import { createProjectMap, validateProjectId } from "../sdk-typescript/src/index.js";
import { type FunctionDraft, type FunctionPublication } from "./index.js";

type Approval = { requester: string; approved: boolean };

export function createApprovalStore() {
  const approvals = new Map<string, Approval>();
  return {
    request(runId: string, requester: string): void {
      if (!runId || !requester || approvals.has(runId))
        throw new Error("invalid approval request");
      approvals.set(runId, { requester, approved: false });
    },
    approve(runId: string, approver: string): boolean {
      const approval = approvals.get(runId);
      if (!approval || approval.approved || !approver || approver === approval.requester)
        return false;
      approval.approved = true;
      return true;
    },
    isApproved(runId: string): boolean {
      return approvals.get(runId)?.approved === true;
    },
  };
}

export function createFunctionDraftStore() {
  const drafts = createProjectMap<FunctionDraft>();
  return {
    save(projectId: string, functionId: string, source: string, savedAt = Date.now()) {
      validateProjectId(projectId);
      if (!functionId || !source || !Number.isFinite(savedAt))
        throw new Error("invalid function draft");
      const draft = { projectId, functionId, source, savedAt };
      drafts.set(projectId, functionId, draft);
      return { ...draft };
    },
    get(projectId: string, functionId: string) {
      const draft = drafts.get(projectId, functionId);
      return draft ? { ...draft } : undefined;
    },
  };
}

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

export function createSharedPackageStore() {
  const packages = createProjectMap<string>();
  return {
    save(projectId: string, packageId: string, source: string): void {
      validateProjectId(projectId);
      if (!packageId || !source) throw new Error("invalid shared package");
      packages.set(projectId, packageId, source);
    },
    load(projectId: string, packageId: string): string {
      const source = packages.get(projectId, packageId);
      if (!source) throw new Error("shared package not found");
      return source;
    },
  };
}

export function publishSharedHelperArtifact(
  projectId: string,
  packageId: string,
  source: string,
  artifactDigest: string,
  toolchain: string,
  publishedAt: number,
  publish: (
    draft: FunctionDraft,
    digest: string,
    compiler: string,
    at: number,
  ) => FunctionPublication,
): FunctionPublication {
  validateProjectId(projectId);
  if (!packageId || !source.includes("package helper"))
    throw new Error("invalid shared helper artifact");
  return publish(
    {
      projectId,
      functionId: packageId,
      source: `package main\n\n// helper: ${source}`,
      savedAt: publishedAt,
    },
    artifactDigest,
    toolchain,
    publishedAt,
  );
}

export function validateStaleInputVersions(
  actual: Readonly<Record<string, number>>,
  expected: Readonly<Record<string, number>>,
): void {
  for (const [name, version] of Object.entries(expected)) {
    if (
      !/^[A-Za-z_][A-Za-z0-9_]*$/.test(name) ||
      !Number.isSafeInteger(version) ||
      version < 0 ||
      actual[name] !== version
    )
      throw new Error("stale input version");
  }
}
