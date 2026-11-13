import { request, type RpcRequest } from "../sdk-typescript/src/index.js";
import { existsSync, mkdtempSync, rmSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";

export const APPROVED_SDK_IMPORTS = ["brain/sdk", "brain/project", "brain/storage"] as const;
export const ENTRY_PACKAGE_FIXTURE = "package main";
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

export function functionRequest(projectId: string, pluginId: string, version: string, operation: "functions.invoke" | "functions.publish", payload: unknown = {}): RpcRequest {
  return request(`functions-${operation}-${projectId}`, { protocol: 1, projectId, pluginId, version, operation, payload });
}
export type FunctionInvocationRequest = { projectId: string; functionId: string; version: string; inputs: Readonly<Record<string, unknown>>; idempotencyKey: string };
export function createInvocationRequest(input: FunctionInvocationRequest): FunctionInvocationRequest {
  if (!input.projectId || !input.functionId || !input.version || !input.idempotencyKey || !input.inputs || typeof input.inputs !== "object") throw new Error("invalid function invocation request");
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
  if (!endpoint.startsWith("https://")) throw new Error("HTTPS endpoint required");
  const response = await fetcher(endpoint, { method: "POST", headers: { "content-type": "application/json" }, body: JSON.stringify(input) });
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
export class InvocationEffects {
  private readonly committed: unknown[] = [];
  begin(): { write: (effect: unknown) => void; commit: (success: boolean) => void } {
    const staged: unknown[] = [];
    let finished = false;
    return {
      write: (effect) => { if (finished) throw new Error("invocation already finished"); staged.push(effect); },
      commit: (success) => { if (finished) throw new Error("invocation already finished"); finished = true; if (success) this.committed.push(...staged); },
    };
  }
  list(): unknown[] { return [...this.committed]; }
}
export type FunctionDraft = { projectId: string; functionId: string; source: string; savedAt: number };
export class FunctionDraftStore {
  private readonly drafts = new Map<string, FunctionDraft>();
  save(projectId: string, functionId: string, source: string, savedAt = Date.now()): FunctionDraft { if (!projectId || !functionId || !source || !Number.isFinite(savedAt)) throw new Error("invalid function draft"); const draft = { projectId, functionId, source, savedAt }; this.drafts.set(`${projectId}:${functionId}`, draft); return { ...draft }; }
  get(projectId: string, functionId: string): FunctionDraft | undefined { const draft = this.drafts.get(`${projectId}:${functionId}`); return draft ? { ...draft } : undefined; }
}
export type FunctionPublication = FunctionDraft & { version: number; artifactDigest: string; toolchain: string; publishedAt: number };
export class FunctionPublicationStore {
  private readonly publications = new Map<string, FunctionPublication>();
  private readonly activeVersions = new Map<string, number>();
  publish(draft: FunctionDraft, artifactDigest: string, toolchain: string, publishedAt = Date.now()): FunctionPublication {
    if (!draft.projectId || !draft.functionId || !draft.source || !/^sha256:[a-f0-9]+$/.test(artifactDigest) || !/^go[0-9]+\.[0-9]+(?:\.[0-9]+)?$/.test(toolchain) || !Number.isFinite(publishedAt)) throw new Error("invalid function publication");
    const prefix = `${draft.projectId}:${draft.functionId}:`;
    const versions = [...this.publications.keys()].filter((key) => key.startsWith(prefix)).map((key) => Number(key.slice(prefix.length)));
    const version = versions.length ? Math.max(...versions) + 1 : 1;
    const publication = { ...draft, version, artifactDigest, toolchain, publishedAt };
    this.publications.set(`${prefix}${version}`, publication);
    this.activeVersions.set(`${draft.projectId}:${draft.functionId}`, version);
    return { ...publication };
  }
  get(projectId: string, functionId: string, version: number): FunctionPublication | undefined {
    const publication = this.publications.get(`${projectId}:${functionId}:${version}`);
    return publication ? { ...publication } : undefined;
  }
  active(projectId: string, functionId: string): FunctionPublication | undefined {
    const version = this.activeVersions.get(`${projectId}:${functionId}`);
    return version === undefined ? undefined : this.get(projectId, functionId, version);
  }
  rollback(projectId: string, functionId: string, version: number): FunctionPublication {
    const publication = this.get(projectId, functionId, version);
    if (!publication) throw new Error("function publication not found");
    this.activeVersions.set(`${projectId}:${functionId}`, version);
    return publication;
  }
}
export class SharedPackageStore {
  private readonly packages = new Map<string, string>();
  save(projectId: string, packageId: string, source: string): void { if (!projectId || !packageId || !source) throw new Error("invalid shared package"); this.packages.set(`${projectId}:${packageId}`, source); }
  load(projectId: string, packageId: string): string { const source = this.packages.get(`${projectId}:${packageId}`); if (!source) throw new Error("shared package not found"); return source; }
}
export type SourceDiagnostic = { line: number; message: string };
export function diagnoseSource(source: string): SourceDiagnostic[] { const diagnostics: SourceDiagnostic[] = []; const lines = source.split("\n"); lines.forEach((line, index) => { if (/import\s+.*\"(os|net|syscall)\"/.test(line)) diagnostics.push({ line: index + 1, message: "unsupported import" }); }); if (!/package\s+main\b/.test(source)) diagnostics.push({ line: 1, message: "package main required" }); return diagnostics; }
