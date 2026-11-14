import { type PluginFunction, type RpcRequest } from "../sdk-typescript/src/index.js";
import { sessionTokenDigest } from "./session-token-digest.js";
import { createAuthState, type AuthState } from "./auth-state.js";
import { AccountSignup } from "./account-signup.js";
import { AccountSignin } from "./account-signin.js";
import { AccountSessionReader } from "./account-session-reader.js";
import { GuestSessionIssuer } from "./guest-session-issuer.js";

export { guestRequest } from "./guest-request.js";
export { signInRequest } from "./sign-in-request.js";
export { signUpRequest } from "./sign-up-request.js";

export { projectUserID } from "./project-user-id.js";

export { sessionTokenDigest } from "./session-token-digest.js";
export { createPKCEChallenge } from "./create-pkce-challenge.js";
export { verifyPKCE } from "./verify-pkce.js";
export { OIDCStateIssuer } from "./oidc-state-issuer.js";
export { OIDCStateConsumer } from "./oidc-state-consumer.js";
export { DeliveryWriter } from "./delivery-writer.js";
export { DeliveryReader } from "./delivery-reader.js";
export { EmailOTPIssuer } from "./email-otp-issuer.js";
export { EmailOTPVerifier } from "./email-otp-verifier.js";
export { SMSOTPIssuer } from "./sms-otp-issuer.js";
export { SMSOTPVerifier } from "./sms-otp-verifier.js";
export { normalizePhoneNumber } from "./normalize-phone-number.js";
export { OTPLimiter } from "./otp-limiter.js";
export { RecoveryTokenIssuer } from "./recovery-token-issuer.js";
export { RecoveryTokenConsumer } from "./recovery-token-consumer.js";

export { AccountRateLimiter } from "./account-rate-limiter.js";

export { ProjectRateLimiter } from "./project-rate-limiter.js";

export { OperatorSessionIssuer } from "./operator-session-issuer.js";
export { OperatorSessionValidator } from "./operator-session-validator.js";
export { OperatorSessionRevoker } from "./operator-session-revoker.js";

export { IdentityLinker } from "./identity-linker.js";

export type { OIDCConfig, OIDCTokenClaims } from "./oidc-types.js";
export { OIDCConfigWriter } from "./oidc-config-writer.js";
export { OIDCConfigReader } from "./oidc-config-reader.js";

export { validateOIDCToken } from "./validate-oidc-token.js";
export { requireRecentAuthentication } from "./require-recent-authentication.js";
export { publicAuthError } from "./public-auth-error.js";
export { validateDeliveryConfig } from "./validate-delivery-config.js";

export { AuthorizationCodeIssuer } from "./authorization-code-issuer.js";
export { AuthorizationCodeConsumer } from "./authorization-code-consumer.js";
export type { AuthorizationCode } from "./authorization-code-types.js";

export { createAuthState } from "./auth-state.js";
export { AccountSignup } from "./account-signup.js";
export { AccountSignin } from "./account-signin.js";
export { AccountSessionReader } from "./account-session-reader.js";
export { AccountSessionRotator } from "./account-session-rotator.js";
export { AccountSessionRevoker } from "./account-session-revoker.js";
export { AccountPasswordResetter } from "./account-password-resetter.js";
export { GuestSessionIssuer } from "./guest-session-issuer.js";
export { GuestAccountUpgrader } from "./guest-account-upgrader.js";
export type AuthOperations = ReturnType<typeof createAuthOperations>;
export function createAuthOperations(state: AuthState = createAuthState()) {
  const signup = new AccountSignup(state); const signin = new AccountSignin(state); const session = new AccountSessionReader(state); const guest = new GuestSessionIssuer(state);
  return { signup: signup.signup.bind(signup), signin: signin.signin.bind(signin), session: session.session.bind(session), guest: guest.guest.bind(guest) };
}

export function functions(service = createAuthOperations()): Readonly<Record<string, PluginFunction>> {
  const payload = (rpc: RpcRequest) => rpc.params.payload as Record<string, string> | undefined;
  return {
    signup: (rpc) => { const p = payload(rpc); return service.signup(rpc.params.projectId, p?.username ?? "", p?.password ?? ""); },
    signin: (rpc) => { const p = payload(rpc); return service.signin(rpc.params.projectId, p?.username ?? p?.provider ?? "", p?.password ?? String(p?.credential ?? "")); },
    session: (rpc) => { const p = payload(rpc); return service.session(rpc.params.projectId, p?.token ?? ""); },
    guest: (rpc) => service.guest(rpc.params.projectId),
  };
}
