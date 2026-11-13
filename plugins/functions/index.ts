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
export function capDiagnostics(lines: readonly string[], maxLines = 20, maxLineLength = 4096): string[] {
  if (!Number.isSafeInteger(maxLines) || maxLines < 1 || !Number.isSafeInteger(maxLineLength) || maxLineLength < 1 || lines.some((line) => line.length > maxLineLength)) throw new Error("diagnostic limit exceeded");
  return [...lines.slice(0, maxLines)];
}
export const APPROVED_WASI_IMPORTS = ["wasi_snapshot_preview1"] as const;
export function validateWasiImports(imports: readonly string[]): void { if (imports.some((name) => !APPROVED_WASI_IMPORTS.includes(name as typeof APPROVED_WASI_IMPORTS[number]))) throw new Error("unexpected WASI import"); }
