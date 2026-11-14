import { validateProjectId } from "../sdk-typescript/src/index.js";

export class SubscriptionOwnerInvalidator {
  constructor(private readonly grants: Set<string>) {}
  invalidate(projectId: string, oldOwnerId: string): void {
    validateProjectId(projectId, "invalid subscription owner");
    if (!oldOwnerId) throw new Error("invalid subscription owner");
    for (const grant of this.grants)
      if (grant.startsWith(`${projectId}:`) && grant.endsWith(`:${oldOwnerId}`))
        this.grants.delete(grant);
  }
}
