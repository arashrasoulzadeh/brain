# Plugin communication architecture

Brain uses a host-owned, language-neutral plugin contract.

## Transports

- Local synchronous calls: JSON-RPC-style JSON messages over Unix sockets.
- External synchronous calls: the same logical contract over gRPC and mTLS.
- Asynchronous work: the PostgreSQL-backed job queue.
- Client realtime traffic: WebSockets only for subscriptions and gameplay.

The transport is replaceable; the authorization, project scope, lifecycle revision, plugin version, request ID, and error model are not.

## Request flow

1. The Go host resolves project and actor identity.
2. The host checks plugin installation, grants, resource scope, lifecycle, and expiry.
3. The host sends a bounded request to the selected plugin version.
4. The plugin returns a bounded result or typed error.
5. The host validates effects and rechecks authorization before commit.

## TypeScript plugins

TypeScript plugins run as separate Node.js workers. They never run inside the Go host process and never receive database credentials, operator tokens, or unrestricted network access. Each version has its own worker and can be started, tested, promoted, drained, rolled back, or stopped independently.

User-authored functions use a stricter sandboxed runtime and must not be treated as trusted TypeScript plugins.
