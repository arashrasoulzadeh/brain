import { validateProjectId } from "../sdk-typescript/src/index.js";

export class SubscriptionGrantStore {
  constructor(private readonly grants: Set<string>) {}
  grant(projectId: string, channel: string, actorId: string): void {
    validateProjectId(projectId, "invalid subscription grant");
    if (!channel || !actorId) throw new Error("invalid subscription grant");
    this.grants.add(`${projectId}:${channel}:${actorId}`);
  }
}
