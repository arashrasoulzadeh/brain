import { createAuthState, type AuthState } from "./auth-state.js";
import { AccountSignup } from "./account-signup.js";
import { AccountSignin } from "./account-signin.js";
import { AccountSessionReader } from "./account-session-reader.js";
import { GuestSessionIssuer } from "./guest-session-issuer.js";
export function createAuthOperations(state: AuthState = createAuthState()) {
  const signup = new AccountSignup(state);
  const signin = new AccountSignin(state);
  const session = new AccountSessionReader(state);
  const guest = new GuestSessionIssuer(state);
  return {
    signup: signup.signup.bind(signup),
    signin: signin.signin.bind(signin),
    session: session.session.bind(session),
    guest: guest.guest.bind(guest),
  };
}
