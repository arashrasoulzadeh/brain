import { createAuthOperations } from "../../plugins/auth/index.js";
import { FileGrantStore, FileAccessReader } from "../../plugins/files/index.js";
import { projectRow, RowPolicyGranter, RowPolicyChecker } from "../../plugins/tables/index.js";

const auth = createAuthOperations();
const dev = auth.guest("dev");
try {
  auth.session("prod", dev.token);
  throw new Error("API crossed project boundary");
} catch (error) {
  if (!(error instanceof Error) || error.message !== "invalid session") throw error;
}
const policyGrants = new Set<string>();
new RowPolicyGranter(policyGrants).grant("dev", "user-1");
if (
  new RowPolicyChecker(policyGrants).allows("prod", "user-1") ||
  Object.hasOwn(projectRow({ id: "1", secret: "x" }, ["id"]), "secret")
)
  throw new Error("direct SQL exposure crossed policy boundary");
const fileGrants = new Set<string>();
new FileGrantStore(fileGrants).grant("dev", "avatar.png", "user-1");
try {
  new FileAccessReader(fileGrants).read("prod", "avatar.png", "user-1");
  throw new Error("file API crossed project boundary");
} catch (error) {
  if (!(error instanceof Error) || error.message !== "file access denied") throw error;
}
