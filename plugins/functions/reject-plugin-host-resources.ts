const HOST_RESOURCE_KEY = /docker|socket|containerd|host[_-]?path/i;

export function rejectPluginHostResources(value: unknown): void {
  if (Array.isArray(value)) {
    for (const entry of value) rejectPluginHostResources(entry);
    return;
  }
  if (!value || typeof value !== "object") return;
  for (const [key, entry] of Object.entries(value)) {
    if (
      HOST_RESOURCE_KEY.test(key) ||
      (typeof entry === "string" && (entry.includes("/var/run/") || entry.includes("/run/")))
    )
      throw new Error("host resources must remain host-owned");
    rejectPluginHostResources(entry);
  }
}
