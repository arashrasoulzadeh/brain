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
