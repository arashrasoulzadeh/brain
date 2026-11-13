export class FileQuotaLedger {
  private readonly reserved = new Map<string, number>();
  constructor(private readonly quotaByProject: ReadonlyMap<string, number>) {}
  reserve(projectId: string, size: number): void {
    if (!Number.isSafeInteger(size) || size < 0) throw new Error("invalid quota amount");
    const next = (this.reserved.get(projectId) ?? 0) + size;
    const quota = this.quotaByProject.get(projectId);
    if (quota === undefined || next > quota) throw new Error("file quota exceeded");
    this.reserved.set(projectId, next);
  }
  release(projectId: string, size: number): void { this.reserved.set(projectId, Math.max(0, (this.reserved.get(projectId) ?? 0) - size)); }
  reservedFor(projectId: string): number { return this.reserved.get(projectId) ?? 0; }
}
