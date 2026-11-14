export type DataFixture = {
  projectId: string;
  collection: "profiles";
  record: { id: string; displayName: string };
};

export function createDataFixture(projectId = "dev"): DataFixture {
  if (!projectId.trim()) throw new Error("fixture project required");
  return {
    projectId,
    collection: "profiles",
    record: { id: "fixture-user", displayName: "Fixture User" },
  };
}
