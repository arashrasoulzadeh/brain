export const APPROVED_SDK_IMPORTS = ["brain/sdk", "brain/project", "brain/storage"] as const;
export function validateSDKImports(imports: readonly string[]): void { if (imports.some((name) => !APPROVED_SDK_IMPORTS.includes(name as typeof APPROVED_SDK_IMPORTS[number]))) throw new Error("unsupported SDK import"); }
export function requireSDKImports(imports: readonly string[], required: readonly string[]): void { validateSDKImports(imports); if (required.some((name) => !imports.includes(name))) throw new Error("required SDK import missing"); }
