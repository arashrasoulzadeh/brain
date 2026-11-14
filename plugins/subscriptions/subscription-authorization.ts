import { validateProjectId } from "../sdk-typescript/src/index.js";

export class SubscriptionAuthorization {
  constructor(private readonly grants: ReadonlySet<string>) {}
  authorize(projectId: string, channel: string, actorId: string): void {
    validateProjectId(projectId, "subscription denied");
    if (!channel || !actorId) throw new Error("subscription denied");
    if (!this.grants.has(`${projectId}:${channel}:${actorId}`))
      throw new Error("subscription denied");
  }
}
