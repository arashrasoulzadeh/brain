export type PrivateBucketConfig = {
  endpoint: string;
  bucket: string;
  region: string;
  publicAccess: false;
};
export type PendingUpload = {
  projectId: string;
  objectId: string;
  contentType: string;
  size: number;
  status: "pending" | "published" | "aborted";
  createdAt?: number;
  etag?: string;
};
