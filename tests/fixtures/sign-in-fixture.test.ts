import { createSignInFixture } from "./sign-in-fixture.js";

const fixture = createSignInFixture();
if (
  fixture.projectId !== "dev" ||
  !fixture.username.includes("@") ||
  !fixture.password ||
  fixture.expectedSessionState !== "authenticated"
)
  throw new Error("shared sign-in fixture is incomplete");

try {
  createSignInFixture(" ");
  throw new Error("invalid fixture project accepted");
} catch (error) {
  if (!(error instanceof Error) || error.message !== "fixture project required") throw error;
}
