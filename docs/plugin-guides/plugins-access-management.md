# Access management

## Developer

Use the host-authenticated actor and exact route capability to evaluate invocation permissions. Rule resources are derived from the route payload, not trusted from a client-supplied identity. Use the generated access rule model to create, read, list and revoke rules. See the [policy contract](../../plugins/access-management/README.md) for rule JSON, principals, wildcard limits and revision requirements.

## Administrator

Installing this plugin opts a project into deny-by-default invocation policy. Prepare rules before enabling traffic; an installed but unavailable provider fails closed. Grant both declared management capabilities only to the appropriate operator boundary. Upgrade exact-version rules before promotion, and verify deny and allow cases for each actor. Removing the plugin deliberately restores legacy behavior.
