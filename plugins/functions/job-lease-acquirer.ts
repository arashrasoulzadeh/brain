import type { JobLeaseState } from "./job-lease-state.js";
export class JobLeaseAcquirer {
  constructor(private readonly state: JobLeaseState) {}
  acquire(jobId: string, workerId: string, now: number, leaseMs: number): boolean {
    if (
      !jobId ||
      !workerId ||
      !Number.isSafeInteger(now) ||
      !Number.isSafeInteger(leaseMs) ||
      leaseMs < 1
    )
      throw new Error("invalid job lease");
    const current = this.state.leases.get(jobId);
    if (current && current.expiresAt > now) return false;
    this.state.leases.set(jobId, { workerId, expiresAt: now + leaseMs });
    return true;
  }
}
