# Plugin authentication context

The production TypeScript worker supplies an immutable, host-derived `context.auth` on every invocation. It does not read identities, roles, permissions or authentication claims from public RPC payloads.

- `auth.actor`: original actor (user, operator, plugin, host or public), with ID and authenticated flag.
- `auth.caller`: effective principal whose access is checked for the current invocation. Delegation and observations retain the original actor for attribution but use the calling plugin's principal for authorization.
- `auth.actor.level` and `auth.caller.level`: authenticated, system or anonymous. It is not MFA or authentication assurance evidence.
- `auth.permissions`: exact capability, operation and resource admitted for this invocation; not a complete account ACL or a reusable grant.
- `auth.can(capability, resource = "")`: exact match against the admitted permission. The registry remains the authority for admission and commit. Never use actor kind alone to bypass permission checks.

Example:

```ts
export function read(context: PluginContext, input: unknown) {
  const auth = context.auth;
  if (!auth?.can("example.read", "record-1")) throw new Error("access denied");
  return { actor: auth.actor.id, admin: auth.caller.kind === "operator" };
}
```

Manifest response and pipeline references support `$auth.actor.id`, `$auth.actor.kind`, `$auth.actor.authenticated`, `$auth.actor.level` and corresponding `$auth.caller.*` references. Only expose identity information intentionally in returned data.

The operator service supplies the authenticated operator; the broker credential supplies an authenticated project user. Direct host calls have a host principal, not an operator. Standalone worker fixtures without a host envelope receive an anonymous context. No passwords, session tokens or MFA secrets are passed. Legacy standalone SDK fixtures may omit auth. Trusted host adapters can call `Registry.InvokeWithUserSession(request, token)` to resolve the current user directly from the installed auth plugin. This requires an enrolled access-management project and an explicit host grant for `auth.identity` (capability `auth.session`), plus the user's own operation/resource grant. Identity is project-scoped and the target never receives the session token. The production HTTP server still has no generic user plugin endpoint; this is a host API, not a new public network surface.

Durable observation delivery restores the publisher identity from trusted queue metadata after restart; it never grants that actor's privileges to a subscriber. The separate function VM receives the same actor/caller/permission shape for authenticated operator drafts and anonymous public calls. Host lifecycle and function events retain the verified originating principal.

The packaged host uses a bounded shared-UID process/thread budget of 512 instead of 64. Linux counts the Node and esbuild threads of every plugin against `RLIMIT_NPROC`; the former limit prevented the third worker from starting. Filesystem, network, CPU, memory, and invocation limits remain unchanged. Production smoke tests start the complete inventory.
