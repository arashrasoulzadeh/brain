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
