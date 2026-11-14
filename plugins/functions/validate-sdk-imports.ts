export const APPROVED_SDK_IMPORTS = ["brain/sdk", "brain/project", "brain/storage"] as const;
export function validateSDKImports(imports: readonly string[]): void { if (imports.some((name) => !APPROVED_SDK_IMPORTS.includes(name as typeof APPROVED_SDK_IMPORTS[number]))) throw new Error("unsupported SDK import"); }
