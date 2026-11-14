import { type RpcRequest, validateProjectId } from "../sdk-typescript/src/index.js";
import { runCappedRetries } from "./run-capped-retries.js";

export { functionRequest } from "./function-request.js";
export { publishSharedHelperArtifact } from "./shared-helper-publisher.js";

export const ENTRY_PACKAGE_FIXTURE =
  'package main\n\nimport "encoding/json"\n\nfunc main() {\n\t_ = json.NewEncoder(nil)\n}\n';
export const PROJECT_HELPER_FIXTURE = "package helper";
export { wasmBuildSettings } from "./wasm-build-settings.js";
export { APPROVED_SDK_IMPORTS, validateSDKImports } from "./validate-sdk-imports.js";
export { requireSDKImports } from "./require-sdk-imports.js";
export { validateSourceProject } from "./validate-source-project.js";
export { validateBuildOverrides } from "./validate-build-overrides.js";
export { validateFunctionLimits } from "./validate-function-limits.js";
export { DisposableCompilerWorkspace } from "./disposable-compiler-workspace.js";
export { compilerEnvironment } from "./compiler-environment.js";
export { validateCompilerEnvironment } from "./validate-compiler-environment.js";
export { type CompilerLimits, validateCompilerLimits } from "./compiler-limits.js";

export type { FunctionInvocationResult, HttpsFetcher } from "./invocation-types.js";
export { validateInvocationResult } from "./validate-invocation-result.js";
export { parseInvocationResult } from "./parse-invocation-result.js";
export { executeFixture } from "./execute-fixture.js";
export { invokeDirectHttps } from "./invoke-direct-https.js";
export { createInvocationRequest } from "./create-invocation-request.js";
export { loadDeclaredInputs } from "./function-input-validation.js";
export { validateStaleInputVersions } from "./validate-stale-input-versions.js";
export { type FunctionWriteIntent, validateWriteIntents } from "./write-intent-validation.js";
export { capDiagnostics } from "./diagnostic-cap.js";
export { APPROVED_WASI_IMPORTS, validateWasiImports } from "./validate-wasi-imports.js";
export { validateSandboxCapabilities } from "./validate-sandbox-capabilities.js";
export { runWithTimeout } from "./run-with-timeout.js";
export { validateLinearMemory } from "./validate-execution-limits.js";
export { validateStartupDuration } from "./validate-startup-duration.js";
export { rejectSandboxEscapeAttempt } from "./reject-sandbox-escape.js";
export { createInvocationEffects } from "./create-invocation-effects.js";
export { createInvocationCommitStore } from "./create-invocation-commit-store.js";
export { retryInvocation } from "./retry-invocation.js";
export type FunctionSchedule = {
  projectId: string;
  functionId: string;
  expression: string;
  dueAt: number;
};
export type EnqueuedFunction = {
  projectId: string;
  functionId: string;
  functionVersion: number;
  enqueuedAt: number;
};
export { createScheduleStore } from "./create-schedule-store.js";
/* legacy schedule operations are composed by createScheduleStore above. */
/*
  add(projectId: string, functionId: string, expression: string, dueAt: number): FunctionSchedule {
    validateProjectId(projectId);
    if (!functionId || expression.trim().split(/\s+/).length !== 5 || !Number.isSafeInteger(dueAt) || dueAt < 0) throw new Error("invalid schedule");
    const schedule = { projectId, functionId, expression: expression.trim(), dueAt };
    this.schedules.push(schedule);
    return { ...schedule };
  }
  due(now: number): FunctionSchedule[] { return this.schedules.filter((schedule) => schedule.dueAt <= now).map((schedule) => ({ ...schedule })); }
  claim(projectId: string, functionId: string, workerId: string, now: number, leaseMs: number): boolean {
    validateProjectId(projectId);
    if (!workerId || !Number.isSafeInteger(now) || !Number.isSafeInteger(leaseMs) || leaseMs < 1) throw new Error("invalid schedule claim");
    const schedule = this.schedules.find((candidate) => candidate.projectId === projectId && candidate.functionId === functionId && candidate.dueAt <= now);
    if (!schedule) return false;
    const key = `${projectId}:${functionId}`;
    const current = this.claims.get(key);
    if (current && current.expiresAt > now) return false;
    this.claims.set(key, { workerId, expiresAt: now + leaseMs });
    return true;
  }
  recordOccurrence(projectId: string, functionId: string, occurrenceUtc: string): boolean {
    validateProjectId(projectId);
    if (!functionId || !/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}Z$/.test(occurrenceUtc) || Number.isNaN(Date.parse(occurrenceUtc))) throw new Error("invalid UTC occurrence");
    const key = `${projectId}:${functionId}:${occurrenceUtc}`;
    if (this.occurrences.has(key)) return false;
    this.occurrences.add(key);
    return true;
  }
  enqueueLatestMissed(projectId: string, functionId: string, now: number): FunctionSchedule | undefined {
    validateProjectId(projectId);
    if (!Number.isSafeInteger(now) || now < 0) throw new Error("invalid schedule time");
    const schedule = this.schedules.filter((candidate) => candidate.projectId === projectId && candidate.functionId === functionId && candidate.dueAt <= now).sort((left, right) => right.dueAt - left.dueAt)[0];
    if (!schedule) return undefined;
    const key = `${projectId}:${functionId}:${schedule.dueAt}`;
    if (this.enqueued.has(key)) return undefined;
    this.enqueued.add(key);
    return { ...schedule };
  }
  enqueueVersion(projectId: string, functionId: string, functionVersion: number, enqueuedAt: number): EnqueuedFunction {
    validateProjectId(projectId);
    if (!functionId || !Number.isSafeInteger(functionVersion) || functionVersion < 1 || !Number.isSafeInteger(enqueuedAt) || enqueuedAt < 0) throw new Error("invalid function enqueue");
    return { projectId, functionId, functionVersion, enqueuedAt };
  }
}*/
export type RecordChangeTrigger = {
  projectId: string;
  collection: string;
  operation: string;
  target: { functionId: string; version: number };
};
export { createRecordChangeTriggerStore } from "./create-record-change-trigger-store.js";
export { createJobLeaseStore } from "./create-job-lease-store.js";
export { commitWithGrant } from "./commit-with-grant.js";
export { runCappedRetries } from "./run-capped-retries.js";
export { runIdempotentWorkflow } from "./run-idempotent-workflow.js";
export { redactRunStatus } from "./redact-run-status.js";
export { type WorkflowNode, validateWorkflowGraph } from "./validate-workflow-graph.js";
export { type WorkflowCondition, evaluateCondition } from "./evaluate-workflow-condition.js";
export { createApprovalStore } from "./create-approval-store.js";
export type FunctionDraft = {
  projectId: string;
  functionId: string;
  source: string;
  savedAt: number;
};
export { createFunctionDraftStore } from "./create-function-draft-store.js";
export type FunctionPublication = FunctionDraft & {
  version: number;
  artifactDigest: string;
  toolchain: string;
  publishedAt: number;
};
export { createFunctionPublicationStore } from "./create-function-publication-store.js";
export { createSharedPackageStore } from "./create-shared-package-store.js";
export { type SourceDiagnostic, diagnoseSource } from "./diagnose-source.js";
export {
  registerProjectWorkflowAction,
  type ProjectWorkflowAction,
} from "./register-project-workflow-action.js";
export { replaceProjectWorkflowStep } from "./replace-project-workflow-step.js";
