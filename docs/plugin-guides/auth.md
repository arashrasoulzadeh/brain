# Authentication

Project accounts and sessions. [Authoritative manifest](../../plugins/auth/plugin.yaml).

## Developer

Use signup/signin with username and password, guest for a guest session, and session/identity with a session token. Durable usernames follow the project identifier grammar and may not use the reserved guest- prefix. Password hashes and session digests are persisted server-side; do not return or log them.

## Administrator

Enable the declared auth operations, grant only necessary access, and use Accounts and sessions to list redacted records or revoke a username's sessions. For enrolled user invocation, grant host auth.identity and the user's target operation; do not confuse operator and user credentials.

## Example payloads

auth.signup / auth.signin: {"username":"player","password":"example-password"}; auth.guest: {}; auth.session / auth.identity: {"token":"<user-token>"}; auth.admin.accounts: {"after":""}; auth.admin.sessions.revoke: {"username":"player"}.

Payload snippets are arguments to the declared operation, not standalone HTTP routes. Use the [HTTP invocation contract](../http-api.md) and exact manifest version/capability. Host-owned context, internal snapshots and grants must not be supplied by public callers.

## Limitations

The transport SDK is not a public authentication gateway or integrated OAuth provider. Keep real credentials out of samples and logs.
