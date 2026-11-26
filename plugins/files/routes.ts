import { createResourceRequest, createOperationRoutes, requiredString } from "../sdk-typescript/src/index.js";

export const fileRequest = createResourceRequest<"files.upload" | "files.download" | "files.metadata" | "files.publish">("files", "objectId");

type FileRecord = { key: string; data: string; published: boolean };

export class FileRouteStore {
  private readonly files = new Map<string, FileRecord>();

  execute(
    operation: "upload" | "download" | "metadata" | "publish",
    key: string,
    data?: string,
  ): unknown {
    const existing = this.files.get(key);
    if (operation === "upload") {
      if (!data) throw new Error("file data is required");
      const record = { key, data, published: false };
      this.files.set(key, record);
      return { key, uploaded: true, published: false };
    }
    if (!existing) throw new Error("file not found");
    if (operation === "publish") {
      existing.published = true;
      return { key, published: true };
    }
    if (operation === "metadata")
      return { key, published: existing.published, size: existing.data.length };
    return { key, data: existing.data, published: existing.published };
  }
}

export const fileRouteStore = new FileRouteStore();

type FileOperation = "upload" | "download" | "metadata" | "publish";
type FileInput = { key?: unknown; data?: unknown; published?: unknown };

const routes = createOperationRoutes(["upload", "download", "metadata", "publish"], executeFileRoute);
export const { upload, download, metadata, publish } = routes.handlers;
export const dispatchRoute = routes.dispatch;

function executeFileRoute(operation: FileOperation, input: unknown): unknown {
  const payload = (input ?? {}) as FileInput;
  const key = requiredString(payload.key, "file key is required");
  if (operation === "upload")
    return fileRouteStore.execute(operation, key, typeof payload.data === "string" ? payload.data : undefined);
  if (operation === "publish" && payload.published === true) return { key, published: true };
  if (typeof payload.data === "string") {
    if (operation === "download")
      return { key, data: payload.data, published: Boolean(payload.published) };
    if (operation === "metadata")
      return { key, published: Boolean(payload.published), size: payload.data.length };
  }
  return fileRouteStore.execute(operation, key);
}
