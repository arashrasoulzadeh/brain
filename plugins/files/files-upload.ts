import { fileRouteStore } from "./file-route-store.js";

export function upload(input: unknown): unknown {
  const payload = (input ?? {}) as { key?: unknown; data?: unknown };
  const key = typeof payload.key === "string" ? payload.key.trim() : "";
  if (!key) throw new Error("file key is required");
  return fileRouteStore.execute(
    "upload",
    key,
    typeof payload.data === "string" ? payload.data : undefined,
  );
}
