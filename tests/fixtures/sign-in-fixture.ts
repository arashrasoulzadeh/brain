export type SignInFixture = {
  projectId: string;
  username: string;
  password: string;
  expectedSessionState: "authenticated";
};

export function createSignInFixture(projectId = "dev"): SignInFixture {
  if (!projectId.trim()) throw new Error("fixture project required");
  return {
    projectId,
    username: "fixture-user@example.test",
    password: "fixture-password-123",
    expectedSessionState: "authenticated",
  };
}
