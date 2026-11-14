export type FunctionFixture = {
  projectId: string;
  functionId: string;
  version: string;
  inputs: { name: string };
  idempotencyKey: string;
};

export function createFunctionFixture(projectId = "dev"): FunctionFixture {
  if (!projectId.trim()) throw new Error("fixture project required");
  return {
    projectId,
    functionId: "hello",
    version: "1.0.0",
    inputs: { name: "Fixture User" },
    idempotencyKey: "fixture-function-1",
  };
}
