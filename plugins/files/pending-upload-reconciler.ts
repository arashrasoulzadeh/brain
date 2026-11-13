import type { PendingUpload } from "./types.js";

export class PendingUploadReconciler {
  constructor(private readonly uploads: Map<string, PendingUpload>) {}
  reconcile(now: number, maxAgeMs: number): number { let count = 0; for (const [key, upload] of this.uploads) if (upload.status === "pending" && now - (upload.createdAt ?? now) > maxAgeMs) { this.uploads.set(key, { ...upload, status: "aborted" }); count += 1; } return count; }
}
