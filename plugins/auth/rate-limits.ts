import { createFixedWindowLimiter, validateProjectId } from "../sdk-typescript/src/index.js";
import { projectUserID } from "./project-user-id.js";

export class AccountRateLimiter {
  private readonly allowKey: (key: string, now: number) => boolean;
  constructor(limit: number, windowMs: number) {
    if (!Number.isInteger(limit) || limit < 1 || windowMs <= 0) throw new Error("invalid account rate limit");
    this.allowKey = createFixedWindowLimiter(limit, windowMs);
  }
  allow(projectId: string, username: string, now: number): boolean {
    return this.allowKey(projectUserID(projectId, username), now);
  }
}

export class ProjectRateLimiter {
  private readonly allowKey: (key: string, now: number) => boolean;
  constructor(limit: number, windowMs: number) {
    if (!Number.isInteger(limit) || limit < 1 || windowMs <= 0) throw new Error("invalid project rate limit");
    this.allowKey = createFixedWindowLimiter(limit, windowMs);
  }
  allow(projectId: string, now: number): boolean {
    return this.allowKey(validateProjectId(projectId), now);
  }
}

export class OTPLimiter {
  private readonly allowKey: (key: string, now: number) => boolean;
  constructor(limit: number, windowMs: number) {
    this.allowKey = createFixedWindowLimiter(limit, windowMs);
  }
  allow(projectId: string, destination: string, now: number): boolean {
    validateProjectId(projectId);
    return this.allowKey(`${projectId}:${destination}`, now);
  }
}
