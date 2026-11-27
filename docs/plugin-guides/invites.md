# Invites

Room invitation lifecycle. [Authoritative manifest](../../plugins/invites/plugin.yaml).

## Developer

Create invitations for a target player in a room with TTL 1–86400 seconds (default 60). Accept by exactly one ID or code and the intended player; single-use and project/room checks apply. Cancel by ID. Inspect redacted metadata, never reusable codes.

## Administrator

Install required Rooms dependency first, enable Invites, grant scoped room operations. Its Invitations extension shows metadata and cancellation. Protect returned invite codes and coordinate target-player authorization before exposing client access.

## Example payloads

invites.create: {"roomId":"r1","targetPlayerId":"a","ttlSeconds":60}; invites.accept: {"roomId":"r1","playerId":"a","inviteId":"<returned-id>"}; invites.inspect: {"roomId":"r1"}; invites.cancel: {"roomId":"r1","inviteId":"<returned-id>"}.

Payload snippets are arguments to the declared operation, not standalone HTTP routes. Use the [HTTP invocation contract](../http-api.md) and exact manifest version/capability. Host-owned context, internal snapshots and grants must not be supplied by public callers.

## Limitations

Operator worker routes are not a public invitation gateway. The target user must be bound by the trusted application, not asserted by a public caller.
