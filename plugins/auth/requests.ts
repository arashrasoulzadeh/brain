import { pluginRequest } from "../sdk-typescript/src/index.js";

export function guestRequest(projectId: string, pluginId: string, version: string) {
  return pluginRequest("auth", projectId, pluginId, version, "auth.guest", {}, "guest");
}

export function signInRequest(
  projectId: string,
  pluginId: string,
  version: string,
  provider: string,
  credential: unknown,
) {
  return pluginRequest(
    "auth",
    projectId,
    pluginId,
    version,
    "auth.signin",
    { provider, credential },
    "signin",
  );
}

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
