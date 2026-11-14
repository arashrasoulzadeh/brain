import { validateSDKImports } from "./validate-sdk-imports.js";
export function requireSDKImports(imports: readonly string[], required: readonly string[]): void {
  validateSDKImports(imports);
  if (required.some((name) => !imports.includes(name)))
    throw new Error("required SDK import missing");
}
