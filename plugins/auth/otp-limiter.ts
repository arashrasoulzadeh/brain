import { validateProjectId } from "../sdk-typescript/src/index.js";
export class OTPLimiter {
  private readonly counts = new Map<string, { startedAt: number; count: number }>();
  constructor(
    private readonly limit: number,
    private readonly windowMs: number,
  ) {}
  allow(projectId: string, destination: string, now: number): boolean {
    validateProjectId(projectId);
    const key = `${projectId}:${destination}`;
    const current = this.counts.get(key);
    if (!current || now - current.startedAt >= this.windowMs) {
      this.counts.set(key, { startedAt: now, count: 1 });
      return true;
    }
    if (current.count >= this.limit) return false;
    current.count += 1;
    return true;
  }
}
