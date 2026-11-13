export class SubscriptionAuthorizer {
  private readonly grants = new Set<string>();
  grant(projectId: string, channel: string, actorId: string): void {
    if (!projectId || !channel || !actorId) throw new Error("invalid subscription grant");
    this.grants.add(`${projectId}:${channel}:${actorId}`);
  }
  authorize(projectId: string, channel: string, actorId: string): void {
    if (!this.grants.has(`${projectId}:${channel}:${actorId}`)) throw new Error("subscription denied");
  }
  invalidateOwner(projectId: string, oldOwnerId: string): void {
    for (const grant of this.grants) if (grant.startsWith(`${projectId}:`) && grant.endsWith(`:${oldOwnerId}`)) this.grants.delete(grant);
  }
}
