export function enforceUploadSize(size: number, maxSize: number): void {
  if (
    !Number.isSafeInteger(size) ||
    !Number.isSafeInteger(maxSize) ||
    size < 0 ||
    maxSize < 0 ||
    size > maxSize
  )
    throw new Error("upload size exceeded");
}
