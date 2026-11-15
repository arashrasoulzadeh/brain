import { fileRouteStore } from "./file-route-store.js";

export function download(input: unknown): unknown {
  const key =
    typeof (input as { key?: unknown } | null)?.key === "string"
      ? (input as { key: string }).key.trim()
      : "";
  if (!key) throw new Error("file key is required");
  if (typeof (input as { data?: unknown } | null)?.data === "string")
    return {
      key,
      data: (input as { data: string }).data,
      published: Boolean((input as { published?: unknown }).published),
    };
  return fileRouteStore.execute("download", key);
}
