import {
  commitWithGrant,
  createInvocationCommitStore,
  retryInvocation,
} from "../../plugins/functions/index.js";

const store = createInvocationCommitStore();
let executions = 0;
const first = await retryInvocation(
  store,
  "crash-1",
  () => {
    executions += 1;
    return { ok: true };
  },
  [{ path: "profiles/name", value: "ada" }],
);
const recovered = await retryInvocation(
  store,
  "crash-1",
  () => {
    executions += 1;
    return { ok: false };
  },
  [{ path: "profiles/name", value: "changed" }],
);
if (first.ok !== recovered.ok || executions !== 1 || store.effects().length !== 1)
  throw new Error("crash recovery was not idempotent");
let granted = true;
if (
  commitWithGrant(
    () => granted,
    () => "committed",
    "functions.write",
  ) !== "committed"
)
  throw new Error("grant commit failed");
granted = false;
try {
  commitWithGrant(
    () => granted,
    () => "unsafe",
    "functions.write",
  );
  throw new Error("revoked grant committed");
} catch (error) {
  if (!(error instanceof Error) || error.message !== "grant revoked before commit") throw error;
}
