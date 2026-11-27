# Authentication

## Developer

Use `auth.signup` and `auth.signin` with username and password. Read a session or identity with `{"token":"<session-token>"}` through the corresponding operation; `auth.guest` establishes a guest session. The trusted project is bound by the host. Never log passwords, tokens or credential payloads. Treat the helper APIs for OTP and OIDC separately from the exposed manifest routes.

## Administrator

Grant only the declared auth capabilities and verify project isolation with distinct accounts in two projects. Keep operator sessions separate from application sessions. Review token expiry, account storage and rate-limit configuration before a production rollout. In-memory fixtures are not a durable identity store; use the supported host adapter and verify recovery.
