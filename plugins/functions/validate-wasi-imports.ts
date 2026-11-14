export const APPROVED_WASI_IMPORTS = ["wasi_snapshot_preview1"] as const;
export function validateWasiImports(imports: readonly string[]): void { if (imports.some((name) => !APPROVED_WASI_IMPORTS.includes(name as typeof APPROVED_WASI_IMPORTS[number]))) throw new Error("unexpected WASI import"); }
