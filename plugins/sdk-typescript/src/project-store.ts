import { validateProjectId } from "./validation.js";

// In-memory partitioning only; callers still require host-authenticated project context.
export function createProjectMap<Value>() {
  const projects = new Map<string, Map<string, Value>>();

  return {
    get(projectId: string, key: string): Value | undefined {
      return projects.get(validateProjectId(projectId))?.get(key);
    },
    set(projectId: string, key: string, value: Value): void {
      validateProjectId(projectId);
      let entries = projects.get(projectId);
      if (!entries) {
        entries = new Map();
        projects.set(projectId, entries);
      }
      entries.set(key, value);
    },
  };
}


/** Project-scoped set insertion; callers enforce authorization before use. */
export function createProjectSet<Value>() {
  const values = createProjectMap<Set<Value>>();
  return (projectId: string, key: string, value: Value): boolean => {
    const seen = values.get(projectId, key) ?? new Set<Value>();
    if (seen.has(value)) return false;
    seen.add(value);
    values.set(projectId, key, seen);
    return true;
  };
}

export function createProjectLeaseStore(error = "invalid job lease") {
  const leases = createProjectMap<{ workerId: string; expiresAt: number }>();

  function validate(projectId: string, key: string, workerId: string, now: number, leaseMs: number) {
    validateProjectId(projectId);
    if (!key || !workerId || !Number.isSafeInteger(now) || !Number.isSafeInteger(leaseMs) ||
        leaseMs < 1 || !Number.isSafeInteger(now + leaseMs)) throw new Error(error);
  }

  return {
    acquire(projectId: string, key: string, workerId: string, now: number, leaseMs: number): boolean {
      validate(projectId, key, workerId, now, leaseMs);
      const current = leases.get(projectId, key);
      if (current && current.expiresAt > now) return false;
      leases.set(projectId, key, { workerId, expiresAt: now + leaseMs });
      return true;
    },
    renew(projectId: string, key: string, workerId: string, now: number, leaseMs: number): boolean {
      validate(projectId, key, workerId, now, leaseMs);
      const current = leases.get(projectId, key);
      if (!current || current.workerId !== workerId || current.expiresAt <= now) return false;
      current.expiresAt = now + leaseMs;
      return true;
    },
  };
}
