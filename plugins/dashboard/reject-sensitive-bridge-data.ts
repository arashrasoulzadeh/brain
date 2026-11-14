const SENSITIVE_KEY = /(?:token|secret|password|credential|api[_-]?key|confirmation|destructive)/i;

export function rejectSensitiveBridgeData(value: unknown): void {
  if (!value || typeof value !== "object") return;
  if (Array.isArray(value)) {
    for (const item of value) rejectSensitiveBridgeData(item);
    return;
  }
  for (const [key, nested] of Object.entries(value)) {
    if (SENSITIVE_KEY.test(key)) throw new Error("invalid bridge message");
    rejectSensitiveBridgeData(nested);
  }
}
