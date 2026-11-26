import { createProjectMap, validateProjectId } from "../sdk-typescript/src/index.js";

type Lease = { workerId: string; expiresAt: number };

export function createJobLeaseStore() {
  const leases = createProjectMap<Lease>();
  return {
    acquire(projectId: string, jobId: string, workerId: string, now: number, leaseMs: number): boolean {
      validateLease(projectId, jobId, workerId, now, leaseMs);
      const current = leases.get(projectId, jobId);
      if (current && current.expiresAt > now) return false;
      leases.set(projectId, jobId, { workerId, expiresAt: now + leaseMs });
      return true;
    },
    renew(projectId: string, jobId: string, workerId: string, now: number, leaseMs: number): boolean {
      validateLease(projectId, jobId, workerId, now, leaseMs);
      const current = leases.get(projectId, jobId);
      if (!current || current.workerId !== workerId || current.expiresAt <= now) return false;
      current.expiresAt = now + leaseMs;
      return true;
    },
  };
}

function validateLease(projectId: string, jobId: string, workerId: string, now: number, leaseMs: number): void {
  validateProjectId(projectId);
  if (!jobId || !workerId || !Number.isSafeInteger(now) || !Number.isSafeInteger(leaseMs) || leaseMs < 1)
    throw new Error("invalid job lease");
}
