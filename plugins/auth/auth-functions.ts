import type { PluginFunction } from "../sdk-typescript/src/index.js";
import { createAuthOperations } from "./create-auth-operations.js";
import { authGuestFunction } from "./auth-guest-function.js";
import { authSessionFunction } from "./auth-session-function.js";
import { authSigninFunction } from "./auth-signin-function.js";
import { authSignupFunction } from "./auth-signup-function.js";

export type AuthOperations = ReturnType<typeof createAuthOperations>;

export function functions(
  service: AuthOperations = createAuthOperations(),
): Readonly<Record<string, PluginFunction>> {
  return {
    signup: authSignupFunction(service),
    signin: authSigninFunction(service),
    session: authSessionFunction(service),
    guest: authGuestFunction(service),
  };
}
