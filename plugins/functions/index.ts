import { request, type RpcRequest } from "../sdk-typescript/src/index.js";

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

export function functionRequest(projectId: string, pluginId: string, version: string, operation: "functions.invoke" | "functions.publish", payload: unknown = {}): RpcRequest {
  return request(`functions-${operation}-${projectId}`, { protocol: 1, projectId, pluginId, version, operation, payload });
}
