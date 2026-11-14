export type SubscriptionFixture = {
  projectId: string;
  channel: "records.profiles";
  ownerId: string;
  event: { operation: "updated"; recordId: string };
};

export function createSubscriptionFixture(projectId = "dev"): SubscriptionFixture {
  if (!projectId.trim()) throw new Error("fixture project required");
  return {
    projectId,
    channel: "records.profiles",
    ownerId: "fixture-user",
    event: { operation: "updated", recordId: "fixture-user" },
  };
}
