import type { PrivateBucketConfig } from "./types.js";

export function configurePrivateBucket(config: Omit<PrivateBucketConfig, "publicAccess">): PrivateBucketConfig {
  if (!config.endpoint || !config.bucket || !config.region) throw new Error("invalid bucket configuration");
  const endpoint = new URL(config.endpoint);
  if (endpoint.protocol !== "https:") throw new Error("private bucket requires HTTPS");
  return { ...config, publicAccess: false };
}
