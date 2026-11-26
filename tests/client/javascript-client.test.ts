import { JavaScriptClient } from "../../clients/javascript/src/index.js";

const calls: unknown[] = [];
const client = new JavaScriptClient(
  "dev",
  async <Response>(operation: string, projectId: string, payload: unknown) => {
    calls.push({ operation, projectId, payload });
    return (
      operation === "auth.signin" ? { token: "token", username: "alice" } : { username: "alice" }
    ) as Response;
  },
);
if ((await client.signIn("alice", "password")).username !== "alice")
  throw new Error("typed client sign-in failed");
if ((await client.session("token")).username !== "alice")
  throw new Error("typed client session failed");
if (calls.length !== 2 || (calls[0] as { projectId: string }).projectId !== "dev")
  throw new Error("typed client did not bind project");

for (const projectId of ["", " ", "../dev", "dev/project", "-dev", "_dev", "dev\n", "dev space", "é"]) {
  let rejected = false;
  try {
    new JavaScriptClient(projectId, async <Response>() => ({} as Response));
  } catch {
    rejected = true;
  }
  if (!rejected) throw new Error("invalid project identity accepted");
}
for (const projectId of ["dev", "A", "9", "dev-01_test"]) {
  new JavaScriptClient(projectId, async <Response>() => ({} as Response));
}
