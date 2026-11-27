# Access Management

Durable authorization rules. [Authoritative manifest](../../plugins/access-management/plugin.yaml).

## Developer

Install deliberately: this is opt-in enforcement, not a role-management system. Rule matching is exact on principal, plugin, version, operation and capability, with an exact resource or `*` for resource-bound routes. An explicit deny overrides allows. Use the actual verified actor ID (for durable auth users it includes the project prefix), not a user-provided claim. The policy is checked both at admission and storage commit.

## Administrator

Prepare the policy migration before installation; an unavailable or ambiguous provider fails closed. Grant access.rules.read and access.rules.manage, enable exactly one compatible provider, then create rules through the generated model form. Revoke using the current revision. Removing the provider restores legacy behavior and is a security-sensitive operation. Review rules for each version upgrade.

## Example payloads

`access.rules.create`: {"id":"read-avatar","fields":{"principal":"user:dev:alice","plugin":"files","version":"0.1.0","operation":"files.download","capability":"files.download","resource":"avatar.png","allowed":true}}. Use access.rules.read/list to get revision and access.rules.revoke with id/revision.

Payload snippets are arguments to the declared operation, not standalone HTTP routes. Use the [HTTP invocation contract](../http-api.md) and exact manifest version/capability. Host-owned context, internal snapshots and grants must not be supplied by public callers.

## Limitations

No role/group or row-predicate policy; domain checks remain necessary. An unenrolled project is not equivalent to an enrolled deny-by-default project. See [policy details](../../plugins/access-management/README.md).
