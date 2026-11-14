const CREDENTIAL_KEY = /token|secret|password|credential|private[_-]?key|api[_-]?key/i;

export function rejectPluginCredentials(value: unknown): void {
  if (Array.isArray(value)) {
    for (const entry of value) rejectPluginCredentials(entry);
    return;
  }
  if (!value || typeof value !== "object") return;
  for (const [key, entry] of Object.entries(value)) {
    if (CREDENTIAL_KEY.test(key)) throw new Error("publish credentials must remain host-owned");
    rejectPluginCredentials(entry);
  }
}
