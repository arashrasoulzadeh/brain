export type JobLeaseState = { leases: Map<string, { workerId: string; expiresAt: number }> };
export function createJobLeaseState(): JobLeaseState {
  return { leases: new Map() };
}
