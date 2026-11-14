import type { JobLeaseState } from "./job-lease-state.js";
export class JobLeaseRenewer {
  constructor(private readonly state: JobLeaseState) {}
  renew(projectId: string, jobId: string, workerId: string, now: number, leaseMs: number): boolean {
    if (!projectId || !Number.isSafeInteger(now) || !Number.isSafeInteger(leaseMs) || leaseMs < 1)
      throw new Error("invalid job lease");
    const current = this.state.leases.get(`${projectId}:${jobId}`);
    if (!current || current.workerId !== workerId || current.expiresAt <= now) return false;
    current.expiresAt = now + leaseMs;
    return true;
  }
}
