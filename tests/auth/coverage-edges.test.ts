import assert from "node:assert/strict";
import { test } from "node:test";
import {
  createAuthState, AccountSignup, AccountSignin, AccountPasswordResetter, GuestAccountUpgrader,
  AccountSessionReader, AccountSessionRotator, AccountSessionRevoker, GuestSessionIssuer,
  OperatorSessionIssuer, IdentityLinker, projectUserID, publicAuthError,
  AccountRateLimiter, ProjectRateLimiter, DeliveryWriter, DeliveryReader, validateDeliveryConfig,
  firstPartySessionCookie, requireRecentAuthentication, EmailOTPIssuer, EmailOTPVerifier,
  SMSOTPIssuer, RecoveryTokenIssuer, normalizePhoneNumber, AuthorizationCodeIssuer,
  AuthorizationCodeConsumer, OIDCConfigReader, OIDCConfigWriter, OIDCStateIssuer,
  OIDCStateConsumer, validateOIDCToken, verifyPKCE, functions, createAuthOperations,
} from "../../plugins/auth/index.js";
import { newSession } from "../../plugins/auth/sessions.js";
import type { AuthorizationCode, OIDCConfig } from "../../plugins/auth/index.js";
import type { OIDCState } from "../../plugins/auth/oidc.js";
import type { OTPChallenge } from "../../plugins/auth/otp.js";
import { pluginRequest } from "../../plugins/sdk-typescript/src/index.js";
import type { PluginContext } from "../../plugins/sdk-typescript/src/index.js";

test("account mutations reject invalid identity and isolate password reset and guest upgrade", async () => {
  assert.throws(() => createAuthState(0), /invalid session ttl/);
  const state = createAuthState(1000);
  const signup = new AccountSignup(state);
  const signin = new AccountSignin(state);
  const reset = new AccountPasswordResetter(state);
  const upgrade = new GuestAccountUpgrader(state);
  assert.throws(() => projectUserID("bad/id", "user"), /invalid project user id/);
  assert.throws(() => projectUserID("alpha", "bad/id"), /invalid project user id/);
  for (const [name, password] of [["", "password"], ["guest-fake", "password"], ["alice", "short"]]) {
    await assert.rejects(signup.signup("alpha", name, password), /invalid signup/);
  }
  await assert.rejects(reset.resetPassword("alpha", "missing", "password"), /account not found/);
  await signup.signup("alpha", "alice", "password");
  await signup.signup("beta", "alice", "password");
  const alice = await signin.signin("alpha", "alice", "password");
  const beta = await signin.signin("beta", "alice", "password");
  const guest = new GuestSessionIssuer(state).guest("alpha");
  await assert.rejects(upgrade.upgradeGuest("alpha", "missing", "other", "password"), /guest session required/);
  await assert.rejects(upgrade.upgradeGuest("beta", guest.token, "other", "password"), /guest session required/);
  await assert.rejects(upgrade.upgradeGuest("alpha", alice.token, "other", "password"), /guest session required/);
  const expired = newSession(state, "alpha", "guest-expired", 0);
  await assert.rejects(upgrade.upgradeGuest("alpha", expired.token, "other", "password"), /guest session required/);
  await assert.rejects(upgrade.upgradeGuest("alpha", guest.token, "alice", "password"), /account already exists/);
  await reset.resetPassword("alpha", "alice", "new-password");
  const reader = new AccountSessionReader(state);
  assert.throws(() => reader.session("alpha", alice.token), /invalid session/);
  assert.equal(reader.session("beta", beta.token).username, "alice");
  assert.equal(reader.session("alpha", guest.token).username, guest.username);
  assert.equal((await upgrade.upgradeGuest("alpha", guest.token, "upgraded", "password")).userId, guest.userId);
  assert.equal(reader.identity("alpha", guest.token).userId, guest.userId);
  const links = new IdentityLinker();
  for (const args of [["", "provider", "subject"], ["id", " ", "subject"], ["id", "provider", " "]] as const) {
    assert.throws(() => links.link(args[0], args[1], args[2]), /invalid identity link/);
  }
  links.link("id", "provider", "subject");
  links.link("id", "provider", "subject");
  assert.equal(publicAuthError(new Error("account not found")), "invalid credentials");
  assert.equal(publicAuthError(new Error("private detail")), "authentication failed");
  assert.equal(publicAuthError("private detail"), "authentication failed");
});

test("session, cookie, rate and delivery boundaries fail closed", () => {
  const state = createAuthState(100);
  const guest = newSession(state, "alpha", "guest-one", 100);
  const rotator = new AccountSessionRotator(state);
  const revoker = new AccountSessionRevoker(state);
  assert.throws(() => rotator.rotate("alpha", "unknown", 101), /invalid session/);
  assert.throws(() => rotator.rotate("beta", guest.token, 101), /invalid session/);
  assert.throws(() => rotator.rotate("alpha", guest.token, 200), /invalid session/);
  assert.throws(() => revoker.logout("alpha", "unknown"), /invalid session/);
  assert.throws(() => revoker.logout("beta", guest.token), /invalid session/);
  assert.throws(() => new OperatorSessionIssuer(new Map()).issue(" "), /operator id required/);
  for (const token of ["bad;token", "bad token"]) assert.throws(() => firstPartySessionCookie(token, 60), /invalid session cookie/);
  for (const ttl of [0, -1, 1.5]) assert.throws(() => firstPartySessionCookie("valid", ttl), /invalid session cookie/);
  assert.throws(() => requireRecentAuthentication(0, 0, 0), /recent authentication required/);
  for (const Limiter of [AccountRateLimiter, ProjectRateLimiter]) {
    for (const [limit, window] of [[1.5, 100], [0, 100], [1, 0]]) {
      assert.throws(() => new Limiter(limit, window), /invalid .* rate limit/);
    }
  }
  const messages = new Map<string, string[]>();
  const writer = new DeliveryWriter(messages);
  const reader = new DeliveryReader(messages);
  assert.equal(reader.read("missing"), undefined);
  assert.throws(() => writer.deliver("", "body"), /invalid test delivery/);
  assert.throws(() => writer.deliver("target", " "), /invalid test delivery/);
  writer.deliver("target", "one");
  writer.deliver("target", "two");
  assert.equal(reader.read("target"), "two");
  validateDeliveryConfig("production", "smtp");
  assert.throws(() => validateDeliveryConfig("", "smtp"), /invalid delivery configuration/);
  assert.throws(() => validateDeliveryConfig("test", ""), /invalid delivery configuration/);
});

test("OTP challenges reject absent, expired, wrong and replayed credentials", () => {
  const records = new Map<string, OTPChallenge>();
  const messages = new Map<string, string[]>();
  const inbox = new DeliveryWriter(messages);
  const reader = new DeliveryReader(messages);
  const issuer = new EmailOTPIssuer(records, "secret", inbox);
  const verifier = new EmailOTPVerifier(records, "secret");
  for (const [address, purpose, ttl] of [["wrong", "login", 10], ["a@example.test", "", 10], ["a@example.test", "login", 0]] as const) {
    assert.throws(() => issuer.issue("alpha", address, purpose, 0, ttl), /invalid email OTP/);
  }
  assert.throws(() => new SMSOTPIssuer(records, "secret", inbox).issue("alpha", "bad", "login", 0, 10), /invalid SMS OTP/);
  const address = "a@example.test";
  issuer.issue("alpha", address, "login", 0, 1);
  const code = reader.read(address)!;
  assert.equal(verifier.verify("beta", address, "login", code, 1), false);
  assert.equal(verifier.verify("alpha", address, "login", code, 1000), false);
  assert.equal(verifier.verify("alpha", address, "login", "not-a-code", 1), false);
  assert.equal(verifier.verify("alpha", address, "login", code, 1), true);
  assert.equal(verifier.verify("alpha", address, "login", code, 1), false);
  const recovery = new RecoveryTokenIssuer(records, "secret", inbox);
  assert.throws(() => recovery.issue("alpha", "bad", 0, 1), /invalid recovery request/);
  assert.throws(() => recovery.issue("alpha", address, 0, 0), /invalid recovery request/);
  assert.equal(normalizePhoneNumber("+1 (234) 567-8901"), "+12345678901");
  assert.throws(() => normalizePhoneNumber("123"), /invalid phone number/);
});

test("OIDC exchanges bind each credential to the expected project, client and redirect", () => {
  const config = { issuer: "https://issuer.example", audience: "aud", clientId: "client", redirectUri: "https://app.example/callback" };
  const configs = new Map<string, OIDCConfig>();
  const writer = new OIDCConfigWriter(configs);
  const reader = new OIDCConfigReader(configs);
  assert.throws(() => reader.get("missing"), /configuration not found/);
  for (const invalid of [{ audience: "" }, { clientId: "" }, { issuer: "http://insecure" }, { redirectUri: "http://insecure" }]) {
    assert.throws(() => writer.set("alpha", { ...config, ...invalid }), /invalid OIDC configuration/);
  }
  writer.set("alpha", config);
  const copy = reader.get("alpha");
  copy.clientId = "changed";
  assert.equal(reader.get("alpha").clientId, "client");
  const codes = new Map<string, AuthorizationCode>();
  const issuer = new AuthorizationCodeIssuer(codes);
  const consumer = new AuthorizationCodeConsumer(codes);
  for (const [client, redirect, user, ttl] of [["", config.redirectUri, "u", 1], ["client", "", "u", 1], ["client", config.redirectUri, "", 1], ["client", config.redirectUri, "u", 0]] as const) {
    assert.throws(() => issuer.issue("alpha", client, redirect, user, 0, ttl), /invalid authorization code/);
  }
  const code = issuer.issue("alpha", "client", config.redirectUri, "u", 0, 1);
  for (const args of [["missing", "alpha", "client", config.redirectUri, 1], [code, "beta", "client", config.redirectUri, 1], [code, "alpha", "other", config.redirectUri, 1], [code, "alpha", "client", "other", 1], [code, "alpha", "client", config.redirectUri, 1000]] as const) {
    assert.throws(() => consumer.consume(args[0], args[1], args[2], args[3], args[4]), /authorization code unavailable/);
  }
  assert.equal(consumer.consume(code, "alpha", "client", config.redirectUri, 1).userId, "u");
  assert.equal(verifyPKCE("v", "v", "plain"), false);
  const states = new Map<string, OIDCState>();
  const state = new OIDCStateIssuer(states).issue("alpha", "client", config.redirectUri);
  const stateConsumer = new OIDCStateConsumer(states);
  for (const args of [[state, "beta", "client", config.redirectUri], [state, "alpha", "other", config.redirectUri], [state, "alpha", "client", "other"]] as const) {
    assert.throws(() => stateConsumer.consume(args[0], args[1], args[2], args[3]), /callback state mismatch/);
  }
  stateConsumer.consume(state, "alpha", "client", config.redirectUri);
  const claims = { issuer: config.issuer, audience: config.audience, expiresAt: 100, nonce: "n", signature: "s" };
  for (const invalid of [{ issuer: "wrong" }, { audience: "wrong" }, { expiresAt: 1 }, { nonce: "wrong" }, { signature: "wrong" }]) {
    assert.throws(() => validateOIDCToken({ ...claims, ...invalid }, config, "n", 1, (s) => s === "s"), /invalid OIDC token/);
  }
});

test("auth dispatch uses host project context and emits only successful observations", async () => {
  const service = functions(createAuthOperations());
  const events: string[] = [];
  const context: PluginContext = {
    project: { id: "alpha", active: true },
    plugin: { id: "auth", version: "1" },
    execution: { operation: "signin" },
    plugins: { observe: { fire(event) { events.push(event); } } },
  };
  const rpc = (operation: string, payload?: unknown, project = "alpha") => pluginRequest("auth", project, "auth", "1", operation, payload);
  await assert.rejects(async () => service.signup(rpc("signup"), context), /invalid signup/);
  await assert.rejects(async () => service.signup(rpc("signup", { username: "alice" }), context), /invalid signup/);
  await service.signup(rpc("signup", { username: "alice", password: "password" }), context);
  await assert.rejects(async () => service.signin(rpc("signin"), context), /invalid credentials/);
  const result = await service.signin(rpc("signin", { provider: "alice", credential: "password" }), context) as { token: string };
  assert.deepEqual(service.session(rpc("session", { token: result.token }), context), { username: "alice" });
  assert.deepEqual(service.identity(rpc("identity", { token: result.token }), context), { userId: "alpha:alice" });
  assert.throws(() => service.session(rpc("session"), context), /invalid session/);
  assert.throws(() => service.identity(rpc("identity"), context), /invalid session/);
  assert.throws(() => service.guest(rpc("guest", {}, "beta"), context), /project context mismatch/);
  service.guest(rpc("guest"), context);
  assert.deepEqual(events, ["account.created", "session.established", "session.established"]);
});
