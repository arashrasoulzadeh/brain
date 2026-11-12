# Phase 03 — App data and authentication

[Phase checklist](../phases.md) · [Architecture plan](../plan.md)

Build on the project, plugin permissions, and guest sessions from Phase 01, exercised in the Phase 02 arena. Deliver production-capable username/password and configured Google sign-in, locally simulated email and SMS flows, native PostgreSQL table access, files, and app subscriptions. Make each feature available through its registered plugin and project grants.

For each slice, run the new check before changing behavior, make it pass through the public entry point, then refactor with the suite green. Keep policy decisions in the host, keep providers and storage adapters behind focused interfaces, and reuse the same project and caller checks across transports.

## Identity and sessions

- [ ] Add failing tests for project-scoped username uniqueness, registration without email delivery, one-time invitation replay, and an expired invitation.
- [ ] Add username/password registration and sign-in with Argon2id, per-account and project rate limits, and manually shared invitation links or codes.
- [ ] Add failing tests for session rotation, refresh-token reuse, logout, and operator revocation. Verify that a session from one project cannot authenticate to another.
- [ ] Complete the shared session flow with token digests, expiry, rotation, and revocation. Apply `Secure`, `HttpOnly`, and `SameSite` cookies, exact allowed origins, and CSRF checks on browser routes.
- [ ] Add failing tests for the hosted native sign-in callback, mismatched state or PKCE verifier, and reuse of an authorization code.
- [ ] Add the native authorization-code exchange, with browser handoff and secure-storage integration exercised from Flutter and Android test clients.
- [ ] Add failing tests for Google OIDC token validation, account-linking proof, and a guest upgrade that preserves the original project user ID. Cover an identity already linked to another account.
- [ ] Add Google OIDC and guest linking through the shared identity and session services. Use the verified issuer and subject to resolve Google identities.
- [ ] Add failing tests for email OTP, phone OTP, local password recovery, expired or consumed codes, and send and guess limits. Exercise phone-number normalization and a reset that revokes existing sessions.
- [ ] Add the local Compose delivery inbox and route email, SMS, and recovery through it in the test profile. Bind challenges to project, destination, and purpose and consume them atomically.
- [ ] Test a production-profile start that rejects fake delivery and exposes username/password, guest, and configured Google independently. Verify that email OTP, SMS OTP, and email recovery stay unavailable until real delivery is configured.

## Native PostgreSQL tables

- [ ] Add failing integration tests with two project schemas and actual migration, direct-writer, and API roles. Verify default denial and rejection of cross-project reads, writes, and pooled caller-context reuse.
- [ ] Provision project schemas and scoped roles. Bind the verified caller transaction-locally and enforce row policies with a non-owner API role.
- [ ] Add failing exposure tests for an unpublished table, a missing primary key, unsupported types, a hidden column, and a filter outside the allowlist.
- [ ] Publish ordinary tables as named resources with a single supported primary key, explicit operation and column grants, and approved filters. Use validated identifiers and parameterized values.
- [ ] Add failing CRUD tests for cursor pagination, invalid values, stale versions, and retried creates. Cover server-only, project-member, record-owner, and explicit public-read policies.
- [ ] Implement bounded CRUD through the same exposure and policy checks. Maintain record versions in PostgreSQL so direct updates also invalidate stale API writes.
- [ ] Add failing tests for dashboard schema changes, external DDL drift, and exposure after the underlying table or trigger changes.
- [ ] Add guided table, column, constraint, and index changes. Provide an advanced SQL migration flow with scoped credentials, preview, recent operator authentication, and an audit entry. Suspend affected exposure until revalidation succeeds.

## Change capture and app subscriptions

- [ ] Add failing tests that commit and roll back API and direct SQL changes. Cover deletes, ownership transfers, unregistered relations, and capture queue saturation.
- [ ] Install the protected PostgreSQL row capture hook. Capture bounded approved metadata and old and new authorization scopes in the originating transaction. Preserve the capture contract for Phase 04 job enqueue.
- [ ] Add failing socket tests for a reused or wrong-endpoint ticket, an unapproved browser origin, unauthorized subscription scope, and revocation during an open connection.
- [ ] Add the app WebSocket endpoint using scoped single-use tickets and per-subscription authorization. Apply independent limits for unauthenticated connections, frames, and send queues.
- [ ] Add failing tests for owner transfer invalidations, writes during the initial fetch, slow consumers, and reconnect with missed events.
- [ ] Route committed changes to authorized invalidation messages. Exercise subscribe-then-fetch and refresh over HTTP after reconnect.

## Files

- [ ] Add failing S3-compatible integration tests for unauthorized reads and deletes, a cross-project object ID, upload limits, quota contention, and an interrupted transfer.
- [ ] Add private object keys, pending metadata, and quota reservation. Stream authorized uploads and downloads through Go and publish metadata only after the object write succeeds.
- [ ] Add failing lifecycle tests for expired pending uploads, orphaned objects, a failed delete, and a disallowed inline content type.
- [ ] Reconcile incomplete writes and deletes. Serve protected downloads as attachments and require an explicit publish action for allowed public assets on their separate origin.

## Dashboard and quickstart

- [ ] Add failing operator tests for project selection, auth and signup settings, exposure and row policies, and file access changes. Verify recent authentication on destructive actions and reject unauthorized form submissions.
- [ ] Extend the Go dashboard with auth settings, record browsing and edits, table management, and file operations. Use audited host actions and sanitized errors. Reuse these actions in the Phase 05 custom UI bridge.
- [ ] Add a browser quickstart that signs in, writes an exposed record, receives an invalidation, and uploads a file. Use the shared API contract to check native sign-in smoke tests.

## Exit gate

- [ ] Run the full suite against local Compose using separate project credentials, real PostgreSQL roles, a private S3-compatible bucket, and the local delivery inbox.
- [ ] Verify production-profile fake-delivery rejection and prove cross-project isolation for the API and direct SQL paths.
- [ ] Re-run the Phase 02 guest co-op flow with the shared session and plugin changes in place.
