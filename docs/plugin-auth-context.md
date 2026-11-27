# Plugin authentication context

The production TypeScript worker supplies an immutable, host-derived `context.auth` on every invocation. It does not read identities, roles, permissions or authentication claims from public RPC payloads.

- `auth.actor`: original actor (user, operator, plugin, host or public), with ID and authenticated flag.
- `auth.caller`: effective principal whose access is checked for the current invocation. Delegation and observations retain the original actor for attribution but use the calling plugin's principal for authorization.
- `auth.level` is represented on each actor/caller: authenticated, system or anonymous. It is not MFA or authentication assurance evidence.
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

The operator service supplies the authenticated operator; the broker credential supplies an authenticated project user. Direct host calls have a host principal, not an operator. Standalone worker fixtures without a host envelope receive an anonymous context. No passwords, session tokens or MFA secrets are passed. Legacy standalone SDK fixtures may omit auth. The current production HTTP server exposes admin and public function routes, not a general authenticated user plugin RPC endpoint.
