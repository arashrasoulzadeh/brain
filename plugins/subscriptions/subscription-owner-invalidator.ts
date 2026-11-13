export class SubscriptionOwnerInvalidator {
  constructor(private readonly grants: Set<string>) {}
  invalidate(projectId: string, oldOwnerId: string): void {
    for (const grant of this.grants) if (grant.startsWith(`${projectId}:`) && grant.endsWith(`:${oldOwnerId}`)) this.grants.delete(grant);
  }
}
