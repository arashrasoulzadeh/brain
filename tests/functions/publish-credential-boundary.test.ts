import { rejectPluginCredentials } from "../../plugins/functions/index.js";

rejectPluginCredentials({ projectId: "dev", artifact: "build.zip" });
for (const input of [
  { publishToken: "secret" },
  { nested: { privateKey: "secret" } },
  { credentials: { username: "user" } },
]) {
  try {
    rejectPluginCredentials(input);
    throw new Error("publish credential crossed plugin boundary");
  } catch (error) {
    if (!(error instanceof Error) || error.message !== "publish credentials must remain host-owned")
      throw error;
  }
}
