import { pluginRequest } from "../sdk-typescript/src/index.js";

export function signUpRequest(
  projectId: string,
  pluginId: string,
  version: string,
  username: string,
  password: string,
) {
  return pluginRequest(
    "auth",
    projectId,
    pluginId,
    version,
    "auth.signup",
    { username, password },
    `signup-${username}`,
  );
}
