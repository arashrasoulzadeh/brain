import { type RpcRequest, validateProjectId } from "../sdk-typescript/src/index.js";
import { existsSync, mkdtempSync, rmSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { isIP } from "node:net";
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
import { createJobLeaseState } from "./job-lease-state.js";
import { JobLeaseAcquirer } from "./job-lease-acquirer.js";
import { JobLeaseRenewer } from "./job-lease-renewer.js";
import { createInvocationEffectsState } from "./invocation-effects-state.js";
import { InvocationEffectsWriter } from "./invocation-effects-writer.js";
import { InvocationEffectsReader } from "./invocation-effects-reader.js";
import { createFunctionPublicationState } from "./function-publication-state.js";
import { FunctionPublicationWriter } from "./function-publication-writer.js";
import { FunctionPublicationReader } from "./function-publication-reader.js";
import { FunctionPublicationRollback } from "./function-publication-rollback.js";
import { createInvocationCommitState } from "./invocation-commit-state.js";
import { InvocationCommitWriter } from "./invocation-commit-writer.js";
import { InvocationCommitReader } from "./invocation-commit-reader.js";
import { InvocationResultReader } from "./invocation-result-reader.js";
import { InvocationExistenceReader } from "./invocation-existence-reader.js";
import { createRecordTriggerState } from "./record-trigger-state.js";
import { RecordTriggerAdder } from "./record-trigger-adder.js";
import { RecordTriggerMatcher } from "./record-trigger-matcher.js";
import { RecordTriggerDispatcher } from "./record-trigger-dispatcher.js";
import { createScheduleState } from "./schedule-state.js";
import { ScheduleAdder } from "./schedule-adder.js";
import { ScheduleReader } from "./schedule-reader.js";
import { ScheduleClaimer } from "./schedule-claimer.js";
import { ScheduleOccurrenceRecorder } from "./schedule-occurrence-recorder.js";
import { MissedScheduleEnqueuer } from "./missed-schedule-enqueuer.js";
import { FunctionVersionEnqueuer } from "./function-version-enqueuer.js";

export { functionRequest } from "./function-request.js";
export { publishSharedHelperArtifact } from "./shared-helper-publisher.js";

export const APPROVED_SDK_IMPORTS = ["brain/sdk", "brain/project", "brain/storage"] as const;
export const ENTRY_PACKAGE_FIXTURE = "package main\n\nimport \"encoding/json\"\n\nfunc main() {\n\t_ = json.NewEncoder(nil)\n}\n";
export const PROJECT_HELPER_FIXTURE = "package helper";
export function wasmBuildSettings(): Readonly<{ goos: "wasip1"; goarch: "wasm"; cgoEnabled: "0" }> {
  return { goos: "wasip1", goarch: "wasm", cgoEnabled: "0" };
}

export function validateSDKImports(imports: readonly string[]): void {
  if (imports.some((name) => !APPROVED_SDK_IMPORTS.includes(name as typeof APPROVED_SDK_IMPORTS[number]))) throw new Error("unsupported SDK import");
}
export function requireSDKImports(imports: readonly string[], required: readonly string[]): void {
  validateSDKImports(imports);
  if (required.some((name) => !imports.includes(name))) throw new Error("required SDK import missing");
}
export function validateSourceProject(sourceProjectId: string, requestedProjectId: string): void {
  if (!sourceProjectId || sourceProjectId !== requestedProjectId) throw new Error("cross-project source access denied");
}
export function validateBuildOverrides(overrides: Readonly<Record<string, string>>): void {
  const denied = new Set(["GOWORK", "GOTOOLCHAIN", "GOENV", "GOMOD", "GOPROXY", "GOFLAGS"]);
  if (Object.keys(overrides).some((key) => denied.has(key) || key.toLowerCase().includes("workspace") || key.toLowerCase().includes("toolchain"))) throw new Error("build override denied");
}
export function validateFunctionLimits(sourceBytes: number, artifactBytes: number, maxSourceBytes: number, maxArtifactBytes: number): void {
  if (![sourceBytes, artifactBytes, maxSourceBytes, maxArtifactBytes].every(Number.isSafeInteger) || sourceBytes < 0 || artifactBytes < 0 || maxSourceBytes < 0 || maxArtifactBytes < 0 || sourceBytes > maxSourceBytes || artifactBytes > maxArtifactBytes) throw new Error("function size limit exceeded");
}
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

export type FunctionInvocationRequest = { projectId: string; functionId: string; version: string; inputs: Readonly<Record<string, unknown>>; idempotencyKey: string };
export function createInvocationRequest(input: FunctionInvocationRequest): FunctionInvocationRequest {
  validateProjectId(input.projectId);
  if (!input.functionId || !input.version || !input.idempotencyKey || !input.inputs || typeof input.inputs !== "object") throw new Error("invalid function invocation request");
  return { ...input, inputs: { ...input.inputs } };
}
export type FunctionInvocationResult = { success: boolean; output?: unknown; error?: string; diagnostics: readonly string[] };
export function validateInvocationResult(result: FunctionInvocationResult, maxDiagnostics = 20): FunctionInvocationResult {
  if (typeof result.success !== "boolean" || result.diagnostics.length > maxDiagnostics || result.diagnostics.some((line) => line.length > 4096) || (result.success && result.error) || (!result.success && !result.error)) throw new Error("invalid function invocation result");
  return { ...result, diagnostics: [...result.diagnostics] };
}
export function parseInvocationResult(value: unknown): FunctionInvocationResult {
  if (!value || typeof value !== "object") throw new Error("malformed function output");
  const result = value as Partial<FunctionInvocationResult>;
  if (Object.keys(value).some((key) => !["success", "output", "error", "diagnostics"].includes(key))) throw new Error("extra function result rejected");
  if (!Array.isArray(result.diagnostics)) throw new Error("malformed function output");
  return validateInvocationResult(result as FunctionInvocationResult);
}
export function executeFixture(fixture: () => unknown): FunctionInvocationResult {
  try {
    return parseInvocationResult(fixture());
  } catch (error) {
    return { success: false, error: error instanceof Error ? error.message : "fixture execution failed", diagnostics: [] };
  }
}
export type HttpsFetcher = (input: string, init?: RequestInit) => Promise<Response>;
export async function invokeDirectHttps(endpoint: string, input: unknown, fetcher: HttpsFetcher = fetch): Promise<FunctionInvocationResult> {
  const target = new URL(endpoint);
  if (target.protocol !== "https:" || target.username || target.password) throw new Error("HTTPS endpoint required");
  const hostname = target.hostname.toLowerCase();
  const addressType = isIP(hostname);
  const privateIPv4 = addressType === 4 && (/^(10|127)\./.test(hostname) || /^192\.168\./.test(hostname) || /^172\.(1[6-9]|2[0-9]|3[0-1])\./.test(hostname) || hostname === "169.254.169.254");
  const privateIPv6 = addressType === 6 && (hostname === "::1" || hostname.startsWith("fc") || hostname.startsWith("fd") || hostname.startsWith("fe8") || hostname.startsWith("fe9") || hostname.startsWith("fea") || hostname.startsWith("feb"));
  if (!hostname || hostname === "localhost" || privateIPv4 || privateIPv6) throw new Error("HTTPS endpoint target denied");
  const response = await fetcher(target.toString(), { method: "POST", redirect: "error", headers: { "content-type": "application/json" }, body: JSON.stringify(input) });
  if (!response.ok) throw new Error(`HTTPS invocation failed: ${response.status}`);
  return parseInvocationResult(await response.json());
}
export function loadDeclaredInputs(inputs: Readonly<Record<string, unknown>>, declarations: readonly string[]): Readonly<Record<string, unknown>> {
  if (!inputs || typeof inputs !== "object" || declarations.some((name) => !/^[A-Za-z_][A-Za-z0-9_]*$/.test(name))) throw new Error("invalid input declaration");
  const loaded: Record<string, unknown> = {};
  for (const name of declarations) {
    if (!Object.prototype.hasOwnProperty.call(inputs, name)) throw new Error("declared input missing");
    loaded[name] = inputs[name];
  }
  return loaded;
}
export type FunctionWriteIntent = { operation: "set" | "delete"; path: string; value?: unknown };
export function validateWriteIntents(intents: readonly FunctionWriteIntent[], declaredPaths: readonly string[]): FunctionWriteIntent[] {
  const declared = new Set(declaredPaths);
  for (const intent of intents) {
    if (!intent || !["set", "delete"].includes(intent.operation) || !intent.path || intent.path.includes("..") || intent.path.startsWith("/") || intent.path.endsWith("/") || intent.path.includes("//")) throw new Error("invalid write intent");
    if (!declared.has(intent.path)) throw new Error("write intent not declared");
  }
  return intents.map((intent) => ({ ...intent }));
}
export function validateStaleInputVersions(actual: Readonly<Record<string, number>>, expected: Readonly<Record<string, number>>): void {
  for (const [name, version] of Object.entries(expected)) {
    if (!/^[A-Za-z_][A-Za-z0-9_]*$/.test(name) || !Number.isSafeInteger(version) || version < 0 || actual[name] !== version) throw new Error("stale input version");
  }
}
export function capDiagnostics(lines: readonly string[], maxLines = 20, maxLineLength = 4096): string[] {
  if (!Number.isSafeInteger(maxLines) || maxLines < 1 || !Number.isSafeInteger(maxLineLength) || maxLineLength < 1 || lines.some((line) => line.length > maxLineLength)) throw new Error("diagnostic limit exceeded");
  return [...lines.slice(0, maxLines)];
}
export const APPROVED_WASI_IMPORTS = ["wasi_snapshot_preview1"] as const;
export function validateWasiImports(imports: readonly string[]): void { if (imports.some((name) => !APPROVED_WASI_IMPORTS.includes(name as typeof APPROVED_WASI_IMPORTS[number]))) throw new Error("unexpected WASI import"); }
export function validateSandboxCapabilities(capabilities: readonly string[]): void {
  const denied = new Set(["filesystem", "network", "environment-credentials"]);
  if (capabilities.some((capability) => denied.has(capability))) throw new Error("sandbox capability denied");
}
export async function runWithTimeout<T>(task: Promise<T>, timeoutMs: number): Promise<T> {
  if (!Number.isSafeInteger(timeoutMs) || timeoutMs < 1) throw new Error("invalid invocation timeout");
  let timer: ReturnType<typeof setTimeout>;
  const timeout = new Promise<never>((_, reject) => { timer = setTimeout(() => reject(new Error("function invocation timed out")), timeoutMs); });
  try { return await Promise.race([task, timeout]); } finally { clearTimeout(timer!); }
}
export function validateLinearMemory(bytes: number, maxBytes: number): void { if (!Number.isSafeInteger(bytes) || !Number.isSafeInteger(maxBytes) || bytes < 0 || maxBytes < 1 || bytes > maxBytes) throw new Error("linear memory limit exceeded"); }
export function validateStartupDuration(durationMs: number, maxDurationMs: number): void { if (!Number.isSafeInteger(durationMs) || !Number.isSafeInteger(maxDurationMs) || durationMs < 0 || maxDurationMs < 1 || durationMs > maxDurationMs) throw new Error("startup timeout exceeded"); }
export function rejectSandboxEscapeAttempt(value: string): void { if (!value || value.includes("../") || value.includes("\\..\\") || /^https?:\/\//.test(value)) throw new Error("sandbox escape attempt rejected"); }
export function createInvocationEffects() { const state = createInvocationEffectsState(); const writer = new InvocationEffectsWriter(state); const reader = new InvocationEffectsReader(state); return { begin: writer.begin.bind(writer), list: reader.list.bind(reader) }; }
export function createInvocationCommitStore() { const state = createInvocationCommitState(); const writer = new InvocationCommitWriter(state); const effects = new InvocationCommitReader(state); const result = new InvocationResultReader(state); const existence = new InvocationExistenceReader(state); return { commit: writer.commit.bind(writer), effects: effects.effects.bind(effects), result: result.result.bind(result), has: existence.has.bind(existence) }; }
export async function retryInvocation<T>(store: ReturnType<typeof createInvocationCommitStore>, idempotencyKey: string, execute: () => T | Promise<T>, effects: readonly unknown[] = []): Promise<T> {
  if (store.has(idempotencyKey)) return store.result(idempotencyKey) as T;
  const result = await execute();
  return store.commit(idempotencyKey, effects, result) as T;
}
export type FunctionSchedule = { projectId: string; functionId: string; expression: string; dueAt: number };
export type EnqueuedFunction = { projectId: string; functionId: string; functionVersion: number; enqueuedAt: number };
export function createScheduleStore() { const state = createScheduleState(); const adder = new ScheduleAdder(state); const reader = new ScheduleReader(state); const claimer = new ScheduleClaimer(state); const occurrence = new ScheduleOccurrenceRecorder(state); const missed = new MissedScheduleEnqueuer(state); const version = new FunctionVersionEnqueuer(); return { add: adder.add.bind(adder), due: reader.due.bind(reader), claim: claimer.claim.bind(claimer), recordOccurrence: occurrence.recordOccurrence.bind(occurrence), enqueueLatestMissed: missed.enqueueLatestMissed.bind(missed), enqueueVersion: version.enqueueVersion.bind(version) }; }
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
export function createRecordChangeTriggerStore() { const state = createRecordTriggerState(); const adder = new RecordTriggerAdder(state); const matcher = new RecordTriggerMatcher(state); const dispatcher = new RecordTriggerDispatcher(state); return { add: adder.add.bind(adder), match: matcher.match.bind(matcher), dispatch: dispatcher.dispatch.bind(dispatcher) }; }
export function createJobLeaseStore() { const state = createJobLeaseState(); const acquirer = new JobLeaseAcquirer(state); const renewer = new JobLeaseRenewer(state); return { acquire: acquirer.acquire.bind(acquirer), renew: renewer.renew.bind(renewer) }; }
export function commitWithGrant<T>(hasGrant: () => boolean, commit: () => T, capability: string): T {
  if (!capability || !hasGrant()) throw new Error("grant revoked before commit");
  return commit();
}
export async function runCappedRetries<T>(task: () => T | Promise<T>, maxAttempts: number): Promise<T> {
  if (!Number.isSafeInteger(maxAttempts) || maxAttempts < 1) throw new Error("invalid retry cap");
  let lastError: unknown;
  for (let attempt = 0; attempt < maxAttempts; attempt++) {
    try { return await task(); } catch (error) { lastError = error; }
  }
  throw lastError instanceof Error ? lastError : new Error("retry attempts exhausted");
}
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
