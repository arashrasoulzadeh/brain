import { existsSync, mkdtempSync, rmSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";

export function compilerEnvironment(): Readonly<Record<string, string>> {
  return { GOOS: "wasip1", GOARCH: "wasm", CGO_ENABLED: "0", GOPROXY: "off", GOSUMDB: "off" };
}

export type CompilerLimits = {
  cpuMs: number;
  memoryBytes: number;
  processes: number;
  wallMs: number;
};
export function validateCompilerLimits(limits: CompilerLimits): void {
  if (
    ![limits.cpuMs, limits.memoryBytes, limits.processes, limits.wallMs].every(
      Number.isSafeInteger,
    ) ||
    limits.cpuMs < 1 ||
    limits.memoryBytes < 1 ||
    limits.processes < 1 ||
    limits.wallMs < 1
  )
    throw new Error("invalid compiler limits");
}

export function validateCompilerEnvironment(environment: Readonly<Record<string, string>>): void {
  const forbidden = /TOKEN|SECRET|PASSWORD|CREDENTIAL|AWS_|HOME|HTTP_PROXY|HTTPS_PROXY|ALL_PROXY/i;
  if (Object.keys(environment).some((key) => forbidden.test(key)))
    throw new Error("compiler credential or network environment denied");
}

export class DisposableCompilerWorkspace {
  readonly path: string;
  constructor(prefix = "brain-function-") {
    this.path = mkdtempSync(join(tmpdir(), prefix));
  }
  dispose(): void {
    if (existsSync(this.path)) rmSync(this.path, { recursive: true, force: true });
  }
}

export type SourceDiagnostic = { line: number; message: string };
export function diagnoseSource(source: string): SourceDiagnostic[] {
  const diagnostics: SourceDiagnostic[] = [];
  const lines = source.split("\n");
  lines.forEach((line, index) => {
    if (/import\s+.*\"(os|net|syscall)\"/.test(line))
      diagnostics.push({ line: index + 1, message: "unsupported import" });
  });
  if (!/package\s+main\b/.test(source))
    diagnostics.push({ line: 1, message: "package main required" });
  return diagnostics;
}

export function capDiagnostics(
  lines: readonly string[],
  maxLines = 20,
  maxLineLength = 4096,
): string[] {
  if (
    !Number.isSafeInteger(maxLines) ||
    maxLines < 1 ||
    !Number.isSafeInteger(maxLineLength) ||
    maxLineLength < 1 ||
    lines.some((line) => line.length > maxLineLength)
  )
    throw new Error("diagnostic limit exceeded");
  return [...lines.slice(0, maxLines)];
}

export function validateBuildOverrides(overrides: Readonly<Record<string, string>>): void {
  const denied = new Set(["GOWORK", "GOTOOLCHAIN", "GOENV", "GOMOD", "GOPROXY", "GOFLAGS"]);
  if (
    Object.keys(overrides).some(
      (key) =>
        denied.has(key) ||
        key.toLowerCase().includes("workspace") ||
        key.toLowerCase().includes("toolchain"),
    )
  )
    throw new Error("build override denied");
}

export function wasmBuildSettings(): Readonly<{ goos: "wasip1"; goarch: "wasm"; cgoEnabled: "0" }> {
  return { goos: "wasip1", goarch: "wasm", cgoEnabled: "0" };
}

export function validateFunctionLimits(
  sourceBytes: number,
  artifactBytes: number,
  maxSourceBytes: number,
  maxArtifactBytes: number,
): void {
  if (
    ![sourceBytes, artifactBytes, maxSourceBytes, maxArtifactBytes].every(Number.isSafeInteger) ||
    sourceBytes < 0 ||
    artifactBytes < 0 ||
    maxSourceBytes < 0 ||
    maxArtifactBytes < 0 ||
    sourceBytes > maxSourceBytes ||
    artifactBytes > maxArtifactBytes
  )
    throw new Error("function size limit exceeded");
}

export function validateLinearMemory(bytes: number, maxBytes: number): void {
  if (
    !Number.isSafeInteger(bytes) ||
    !Number.isSafeInteger(maxBytes) ||
    bytes < 0 ||
    maxBytes < 1 ||
    bytes > maxBytes
  )
    throw new Error("linear memory limit exceeded");
}

export function validateStartupDuration(durationMs: number, maxDurationMs: number): void {
  if (
    !Number.isSafeInteger(durationMs) ||
    !Number.isSafeInteger(maxDurationMs) ||
    durationMs < 0 ||
    maxDurationMs < 1 ||
    durationMs > maxDurationMs
  )
    throw new Error("startup timeout exceeded");
}

export const APPROVED_WASI_IMPORTS = ["wasi_snapshot_preview1"] as const;
export function validateWasiImports(imports: readonly string[]): void {
  if (
    imports.some(
      (name) => !APPROVED_WASI_IMPORTS.includes(name as (typeof APPROVED_WASI_IMPORTS)[number]),
    )
  )
    throw new Error("unexpected WASI import");
}

export const APPROVED_SDK_IMPORTS = ["brain/sdk", "brain/project", "brain/storage"] as const;
export function validateSDKImports(imports: readonly string[]): void {
  if (
    imports.some(
      (name) => !APPROVED_SDK_IMPORTS.includes(name as (typeof APPROVED_SDK_IMPORTS)[number]),
    )
  )
    throw new Error("unsupported SDK import");
}

export function requireSDKImports(imports: readonly string[], required: readonly string[]): void {
  validateSDKImports(imports);
  if (required.some((name) => !imports.includes(name)))
    throw new Error("required SDK import missing");
}
