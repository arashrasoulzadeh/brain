import { type RpcRequest, validateProjectId } from "../sdk-typescript/src/index.js";
import { existsSync, mkdtempSync, rmSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { createApprovalState } from "./approval-state.js";
import { ApprovalRequester } from "./approval-requester.js";
import { ApprovalGranter } from "./approval-granter.js";
import { ApprovalReader } from "./approval-reader.js";
import { createSharedPackageState } from "./shared-package-state.js";
import { SharedPackageWriter } from "./shared-package-writer.js";
import { SharedPackageReader } from "./shared-package-reader.js";
import { createFunctionDraftState } from "./function-draft-state.js";
import { FunctionDraftWriter } from "./function-draft-writer.js";
import { FunctionDraftReader } from "./function-draft-reader.js";
import { createFunctionPublicationState } from "./function-publication-state.js";
import { FunctionPublicationWriter } from "./function-publication-writer.js";
import { FunctionPublicationReader } from "./function-publication-reader.js";
import { FunctionPublicationRollback } from "./function-publication-rollback.js";
import type { FunctionInvocationRequest } from "./invocation-types.js";
import { runCappedRetries } from "./run-capped-retries.js";
import { createInvocationCommitStore } from "./create-invocation-commit-store.js";

export { functionRequest } from "./function-request.js";
export { publishSharedHelperArtifact } from "./shared-helper-publisher.js";

export const ENTRY_PACKAGE_FIXTURE = "package main\n\nimport \"encoding/json\"\n\nfunc main() {\n\t_ = json.NewEncoder(nil)\n}\n";
export const PROJECT_HELPER_FIXTURE = "package helper";
export { wasmBuildSettings } from "./wasm-build-settings.js";
export { APPROVED_SDK_IMPORTS, validateSDKImports, requireSDKImports } from "./validate-sdk-imports.js";
export { validateSourceProject } from "./validate-source-project.js";
export { validateBuildOverrides } from "./validate-build-overrides.js";
export { validateFunctionLimits } from "./validate-function-limits.js";
export class DisposableCompilerWorkspace {
  readonly path: string;
  constructor(prefix = "brain-function-") { this.path = mkdtempSync(join(tmpdir(), prefix)); }
  dispose(): void { if (existsSync(this.path)) rmSync(this.path, { recursive: true, force: true }); }
}
export function compilerEnvironment(): Readonly<Record<string, string>> { return { GOOS: "wasip1", GOARCH: "wasm", CGO_ENABLED: "0", GOPROXY: "off", GOSUMDB: "off" }; }
export function validateCompilerEnvironment(environment: Readonly<Record<string, string>>): void {
  const forbidden = /TOKEN|SECRET|PASSWORD|CREDENTIAL|AWS_|HOME|HTTP_PROXY|HTTPS_PROXY|ALL_PROXY/i;
  if (Object.keys(environment).some((key) => forbidden.test(key))) throw new Error("compiler credential or network environment denied");
}
export type CompilerLimits = { cpuMs: number; memoryBytes: number; processes: number; wallMs: number };
export function validateCompilerLimits(limits: CompilerLimits): void {
  if (![limits.cpuMs, limits.memoryBytes, limits.processes, limits.wallMs].every(Number.isSafeInteger) || limits.cpuMs < 1 || limits.memoryBytes < 1 || limits.processes < 1 || limits.wallMs < 1) throw new Error("invalid compiler limits");
}

export type { FunctionInvocationResult, HttpsFetcher } from "./invocation-types.js";
export { validateInvocationResult } from "./validate-invocation-result.js";
export { parseInvocationResult } from "./parse-invocation-result.js";
export { executeFixture } from "./execute-fixture.js";
export { invokeDirectHttps } from "./invoke-direct-https.js";
export function createInvocationRequest(input: FunctionInvocationRequest): FunctionInvocationRequest {
  validateProjectId(input.projectId);
  if (!input.functionId || !input.version || !input.idempotencyKey || !input.inputs || typeof input.inputs !== "object") throw new Error("invalid function invocation request");
  return { ...input, inputs: { ...input.inputs } };
}
export { loadDeclaredInputs, validateStaleInputVersions } from "./function-input-validation.js";
export { type FunctionWriteIntent, validateWriteIntents } from "./write-intent-validation.js";
export { capDiagnostics } from "./diagnostic-cap.js";
export { APPROVED_WASI_IMPORTS, validateWasiImports } from "./validate-wasi-imports.js";
export { validateSandboxCapabilities } from "./validate-sandbox-capabilities.js";
export { runWithTimeout } from "./run-with-timeout.js";
export { validateLinearMemory, validateStartupDuration } from "./validate-execution-limits.js";
export { rejectSandboxEscapeAttempt } from "./reject-sandbox-escape.js";
export { createInvocationEffects } from "./create-invocation-effects.js";
export { createInvocationCommitStore } from "./create-invocation-commit-store.js";
export async function retryInvocation<T>(store: ReturnType<typeof createInvocationCommitStore>, idempotencyKey: string, execute: () => T | Promise<T>, effects: readonly unknown[] = []): Promise<T> {
  if (store.has(idempotencyKey)) return store.result(idempotencyKey) as T;
  const result = await execute();
  return store.commit(idempotencyKey, effects, result) as T;
}
export type FunctionSchedule = { projectId: string; functionId: string; expression: string; dueAt: number };
export type EnqueuedFunction = { projectId: string; functionId: string; functionVersion: number; enqueuedAt: number };
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
export type RecordChangeTrigger = { projectId: string; collection: string; operation: string; target: { functionId: string; version: number } };
export { createRecordChangeTriggerStore } from "./create-record-change-trigger-store.js";
export { createJobLeaseStore } from "./create-job-lease-store.js";
export { commitWithGrant } from "./commit-with-grant.js";
export { runCappedRetries } from "./run-capped-retries.js";
export async function runIdempotentWorkflow<T>(store: ReturnType<typeof createInvocationCommitStore>, idempotencyKey: string, task: () => T | Promise<T>, maxAttempts: number): Promise<T> {
  if (store.has(idempotencyKey)) return store.result(idempotencyKey) as T;
  const result = await runCappedRetries(task, maxAttempts);
  return store.commit(idempotencyKey, [], result) as T;
}
export function redactRunStatus<T>(status: T): T {
  const secret = /TOKEN|SECRET|PASSWORD|CREDENTIAL|API[_-]?KEY/i;
  const visit = (value: unknown): unknown => {
    if (Array.isArray(value)) return value.map(visit);
    if (!value || typeof value !== "object") return value;
    const result: Record<string, unknown> = {};
    for (const [key, nested] of Object.entries(value)) result[key] = secret.test(key) ? "[redacted]" : visit(nested);
    return result;
  };
  return visit(status) as T;
}
export type WorkflowNode = { id: string; next: readonly string[] };
export function validateWorkflowGraph(nodes: readonly WorkflowNode[]): void {
  const byId = new Map(nodes.map((node) => [node.id, node]));
  if (byId.size !== nodes.length || nodes.some((node) => !node.id || node.next.some((next) => !byId.has(next)))) throw new Error("invalid workflow graph");
  const visiting = new Set<string>();
  const visited = new Set<string>();
  const visit = (id: string): void => {
    if (visiting.has(id)) throw new Error("workflow cycle");
    if (visited.has(id)) return;
    visiting.add(id);
    for (const next of byId.get(id)!.next) visit(next);
    visiting.delete(id);
    visited.add(id);
  };
  for (const node of nodes) visit(node.id);
}
export type WorkflowCondition = { field: string; operator: string; value: unknown };
export function evaluateCondition(condition: WorkflowCondition, context: Readonly<Record<string, unknown>>): boolean {
  if (!condition.field || condition.operator !== "equals") throw new Error("unsupported condition");
  return Object.prototype.hasOwnProperty.call(context, condition.field) && Object.is(context[condition.field], condition.value);
}
export function createApprovalStore() { const state = createApprovalState(); const requester = new ApprovalRequester(state); const granter = new ApprovalGranter(state); const reader = new ApprovalReader(state); return { request: requester.request.bind(requester), approve: granter.approve.bind(granter), isApproved: reader.isApproved.bind(reader) }; }
export type FunctionDraft = { projectId: string; functionId: string; source: string; savedAt: number };
export function createFunctionDraftStore() { const state = createFunctionDraftState(); const writer = new FunctionDraftWriter(state); const reader = new FunctionDraftReader(state); return { save: writer.save.bind(writer), get: reader.get.bind(reader) }; }
export type FunctionPublication = FunctionDraft & { version: number; artifactDigest: string; toolchain: string; publishedAt: number };
export function createFunctionPublicationStore() { const state = createFunctionPublicationState(); const writer = new FunctionPublicationWriter(state); const reader = new FunctionPublicationReader(state); const rollback = new FunctionPublicationRollback(state); return { publish: writer.publish.bind(writer), get: reader.get.bind(reader), active: (projectId: string, functionId: string) => { const version = state.activeVersions.get(`${projectId}:${functionId}`); return version === undefined ? undefined : reader.get(projectId, functionId, version); }, rollback: rollback.rollback.bind(rollback) }; }
export function createSharedPackageStore() { const state = createSharedPackageState(); const writer = new SharedPackageWriter(state); const reader = new SharedPackageReader(state); return { save: writer.save.bind(writer), load: reader.load.bind(reader) }; }
export type SourceDiagnostic = { line: number; message: string };
export function diagnoseSource(source: string): SourceDiagnostic[] { const diagnostics: SourceDiagnostic[] = []; const lines = source.split("\n"); lines.forEach((line, index) => { if (/import\s+.*\"(os|net|syscall)\"/.test(line)) diagnostics.push({ line: index + 1, message: "unsupported import" }); }); if (!/package\s+main\b/.test(source)) diagnostics.push({ line: 1, message: "package main required" }); return diagnostics; }
