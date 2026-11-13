import { AuthService } from "../../plugins/auth/index.js";
import { FileGrantStore, FileAccessReader } from "../../plugins/files/index.js";
import { projectRow, RowPolicy } from "../../plugins/tables/index.js";

const auth = new AuthService();
const dev = auth.guest("dev");
try { auth.session("prod", dev.token); throw new Error("API crossed project boundary"); } catch (error) { if (!(error instanceof Error) || error.message !== "invalid session") throw error; }
const policy = new RowPolicy(); policy.grant("dev", "user-1");
if (policy.allows("prod", "user-1") || Object.hasOwn(projectRow({ id: "1", secret: "x" }, ["id"]), "secret")) throw new Error("direct SQL exposure crossed policy boundary");
const fileGrants = new Set<string>(); new FileGrantStore(fileGrants).grant("dev", "avatar.png", "user-1");
try { new FileAccessReader(fileGrants).read("prod", "avatar.png", "user-1"); throw new Error("file API crossed project boundary"); } catch (error) { if (!(error instanceof Error) || error.message !== "file access denied") throw error; }
