# Subscriptions

Bounded observation polling. [Authoritative manifest](../../plugins/subscriptions/plugin.yaml).

## Developer

Open the rooms.connected channel; preserve the returned ticket and cursor. Receive with ticket/channel/cursor, advance the cursor, resync on retention errors and close the ticket when finished. The retained stream has 256 events.

## Administrator

Grant the room observation capability and channel-specific operations. The Realtime subscriptions page exposes redacted active-ticket counts and revocation. Do not log ticket strings.

## Example payloads

subscriptions.open: {"channel":"rooms.connected"}; subscriptions.receive: {"channel":"rooms.connected","ticket":"<returned-ticket>","cursor":0}; subscriptions.resync / subscriptions.close: {"channel":"rooms.connected","ticket":"<returned-ticket>"}; subscriptions.inspect / subscriptions.revoke: {"channel":"rooms.connected"}.

Payload snippets are arguments to the declared operation, not standalone HTTP routes. Use the [HTTP invocation contract](../http-api.md) and exact manifest version/capability. Host-owned context, internal snapshots and grants must not be supplied by public callers.

## Limitations

Only the rooms.connected polling channel is wired. This operator adapter is not user-authorized WebSocket delivery or an unbounded replay log.
