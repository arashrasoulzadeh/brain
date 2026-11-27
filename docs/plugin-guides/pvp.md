# PvP

Shared turn and revision primitives. [Authoritative manifest](../../plugins/pvp/plugin.yaml).

## Developer

Reuse CommandDeduplicator, nextTurn, commitRevision, replayHistory, expireTurn, pauseTurn, resumeTurn and privatePlayerView. The shared SDK is the import boundary used by game workers. Bind membership and project identity in the host.

## Administrator

Grant pvp.describe for read-only discovery. The current route does not manage active matches. Validate version compatibility before changing game command semantics.

## Example payloads

pvp.describe: {} returns projectId, turnBased and revisionedCommands. See [SDK reference](../sdk-reference.md).

Payload snippets are arguments to the declared operation, not standalone HTTP routes. Use the [HTTP invocation contract](../http-api.md) and exact manifest version/capability. Host-owned context, internal snapshots and grants must not be supplied by public callers.

## Limitations

Pure helpers are not durable cross-process locks, trusted actor identity or a network service.
