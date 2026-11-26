import { validateProjectId } from "../sdk-typescript/src/index.js";

export class QuotaReader {
  constructor(private readonly reserved: ReadonlyMap<string, number>) {}
  reservedFor(projectId: string): number {
    validateProjectId(projectId);
    return this.reserved.get(projectId) ?? 0;
  }
}

export class QuotaReleaser {
  constructor(private readonly reserved: Map<string, number>) {}
  release(projectId: string, size: number): void {
    validateProjectId(projectId);
    if (!Number.isSafeInteger(size) || size < 0) throw new Error("invalid quota amount");
    this.reserved.set(projectId, Math.max(0, (this.reserved.get(projectId) ?? 0) - size));
  }
}

export class QuotaReserver {
  constructor(
    private readonly reserved: Map<string, number>,
    private readonly quotaByProject: ReadonlyMap<string, number>,
  ) {}
  reserve(projectId: string, size: number): void {
    validateProjectId(projectId);
    if (!Number.isSafeInteger(size) || size < 0) throw new Error("invalid quota amount");
    const next = (this.reserved.get(projectId) ?? 0) + size;
    const quota = this.quotaByProject.get(projectId);
    if (quota === undefined || next > quota) throw new Error("file quota exceeded");
    this.reserved.set(projectId, next);
  }
}
