import { validateProjectId } from "../sdk-typescript/src/index.js";

export class QuotaReleaser {
  constructor(private readonly reserved: Map<string, number>) {}
  release(projectId: string, size: number): void {
    validateProjectId(projectId);
    if (!Number.isSafeInteger(size) || size < 0) throw new Error("invalid quota amount");
    this.reserved.set(projectId, Math.max(0, (this.reserved.get(projectId) ?? 0) - size));
  }
}
