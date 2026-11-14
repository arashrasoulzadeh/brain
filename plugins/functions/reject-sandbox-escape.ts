export function rejectSandboxEscapeAttempt(value: string): void {
  let decoded = value;
  try {
    decoded = decodeURIComponent(value);
  } catch {
    throw new Error("sandbox escape attempt rejected");
  }
  const normalized = decoded.replaceAll("\\", "/");
  if (
    !value ||
    decoded.includes("\u0000") ||
    normalized.split("/").includes("..") ||
    normalized.startsWith("/") ||
    /^[A-Za-z]:\//.test(normalized) ||
    /^https?:\/\//.test(decoded)
  )
    throw new Error("sandbox escape attempt rejected");
}
