import { fileRouteStore } from "./file-route-store.js";

export function metadata(input: unknown): unknown {
  const key =
    typeof (input as { key?: unknown } | null)?.key === "string"
      ? (input as { key: string }).key.trim()
      : "";
  if (!key) throw new Error("file key is required");
  return fileRouteStore.execute("metadata", key);
}
