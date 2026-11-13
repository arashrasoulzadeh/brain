import { request } from "../sdk-typescript/src/index.js";

export function fileRequest(projectId: string, pluginId: string, version: string, operation: "files.upload" | "files.download" | "files.metadata" | "files.publish", objectId: string, payload: unknown = {}) {
  return request(`files-${operation}-${projectId}-${objectId}`, { protocol: 1, projectId, pluginId, version, operation, payload: { objectId, ...(payload as Record<string, unknown>) } });
}

export type PrivateBucketConfig = { endpoint: string; bucket: string; region: string; publicAccess: false };
export function configurePrivateBucket(config: Omit<PrivateBucketConfig, "publicAccess">): PrivateBucketConfig {
  if (!config.endpoint || !config.bucket || !config.region) throw new Error("invalid bucket configuration");
  const endpoint = new URL(config.endpoint);
  if (endpoint.protocol !== "https:") throw new Error("private bucket requires HTTPS");
  return { ...config, publicAccess: false };
}

