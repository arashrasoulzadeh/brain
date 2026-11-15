import { fileRouteStore } from "./file-route-store.js";

export function publish(input: unknown): unknown {
  const key =
    typeof (input as { key?: unknown } | null)?.key === "string"
      ? (input as { key: string }).key.trim()
      : "";
  if (!key) throw new Error("file key is required");
  if ((input as { published?: unknown } | null)?.published === true)
    return { key, published: true };
  return fileRouteStore.execute("publish", key);
}
