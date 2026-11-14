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
