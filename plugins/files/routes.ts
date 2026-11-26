import { createOperationRoutes, requiredString } from "../sdk-typescript/src/index.js";
import { fileRouteStore } from "./file-route-store.js";

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
