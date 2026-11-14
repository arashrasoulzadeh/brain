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
