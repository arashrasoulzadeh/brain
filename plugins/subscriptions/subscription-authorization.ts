export class SubscriptionAuthorization {
  constructor(private readonly grants: ReadonlySet<string>) {}
  authorize(projectId: string, channel: string, actorId: string): void {
    if (!this.grants.has(`${projectId}:${channel}:${actorId}`)) throw new Error("subscription denied");
  }
}
