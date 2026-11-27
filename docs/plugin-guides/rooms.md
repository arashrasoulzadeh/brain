# Rooms

Room lifecycle and membership. [Authoritative manifest](../../plugins/rooms/plugin.yaml).

## Developer

Use id for room identity, member for joining/reconnecting, and result for a result write. Reading rooms.result without result retrieves existing state. Room-connected observations contain roomId and reconnected. The host supplies durable membership and result snapshots.

## Administrator

Use Room details and Manage rooms in Live sessions; grant room-scoped operations separately. Matchmaking and Invitations may contribute tabs. Inspect dependency and observation impact before stopping Rooms.

## Example payloads

rooms.create: {"id":"r1"}; rooms.join / rooms.reconnect: {"id":"r1","member":"a"}; rooms.result: {"id":"r1","result":{"score":1}} or {"id":"r1"} for a read.

Payload snippets are arguments to the declared operation, not standalone HTTP routes. Use the [HTTP invocation contract](../http-api.md) and exact manifest version/capability. Host-owned context, internal snapshots and grants must not be supplied by public callers.

## Limitations

Rooms owns generic membership, not game rules. Operator routes do not authorize public players or implement a realtime socket endpoint.

Room inventory shows stored open/closed status and member count. Confirm `rooms.admin.remove-member` with an exact room ID and member, or `rooms.admin.close` with an exact room ID. Both require the new `rooms.admin.manage` grant and applicable access rules. Closing preserves the stored result and membership for inspection but blocks joins, reconnects, result writes and recreating the closed ID. Removal edits durable membership, not a realtime socket connection or a permanent ban; the existing join operation may admit the member again while open.
