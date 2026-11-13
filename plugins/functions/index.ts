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

export function functionRequest(projectId: string, pluginId: string, version: string, operation: "functions.invoke" | "functions.publish", payload: unknown = {}): RpcRequest {
  return request(`functions-${operation}-${projectId}`, { protocol: 1, projectId, pluginId, version, operation, payload });
}
