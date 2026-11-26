# Access Management

A project opts in by installing this plugin and granting `access.rules.read` and `access.rules.manage`, then enabling it. An installed but disabled, unready, draining, incompatible, ambiguous, or storage-unavailable policy provider denies feature invocations. Existing projects without the plugin retain their legacy authorization. Installing it before preparing rules intentionally denies feature traffic; schedule migration accordingly.

Only an authenticated operator may manage rules, through the regular operator/admin invocation boundary. Management still requires the installed plugin's capabilities and lifecycle checks. Rules are stored in the project's host-owned model; an operator can create, read, list and revoke them using the generated admin page or API. Revocation uses the current record revision.

An example create payload:

```json
{
  "id": "alice-files-read",
  "fields": {
    "principal": "user:alice",
    "plugin": "files",
    "version": "0.1.0",
    "operation": "files.download",
    "capability": "files.download",
    "resource": "avatar.png",
    "allowed": true
  }
}
```

Principals are exact: `user:<actor-id>` from a host-issued credential, `operator:<actor-id>` from verified operator sessions, `plugin:<id>@<version>` for declared dependency calls and observations, or `host` for trusted internal calls. For example, an analytics observer of rooms.connected uses principal `plugin:rooms@0.1.0`, plugin `analytics`, operation `@observe/rooms/connected`, capability `observe.rooms.connected`, and resource `rooms.connected`.

Rules match exact plugin ID, version, operation and capability. A manifest can declare a `resource_field` for an operation; the host derives that resource directly from the actual payload. A rule uses an exact resource ID or `*` for every resource of that exact operation. Unscoped routes require `resource: ""`; host models use their model name, and observers use the source plugin/event name. Missing rules deny; an explicit deny overrides exact or resource-wide allows. No principal, project, operation or version wildcard is supported. The policy does not replace host lifecycle checks, authentication, resource grants, room membership, file ownership, or other domain invariants.

The registry checks rules before execution and again before committing host-managed effects. Broker credential issue and consumption also check them. New user-facing host integrations should use `Registry.InvokeWithCredential` rather than the trusted internal invocation method.

Access rules apply to every registry-routed plugin automatically; new plugins need no hard-coded permission wiring. Version upgrades require new rules for the new version. For a policy provider upgrade, ensure exactly one ready, enabled compatible version is present. Removing the plugin deliberately restores legacy behavior, so removal must remain operator-controlled.

Existing domain-specific ownership and membership checks remain in place. Models use collection scope unless the host exposes a more specific resource; row-level predicates and role/group management are outside this invocation-policy contract.
