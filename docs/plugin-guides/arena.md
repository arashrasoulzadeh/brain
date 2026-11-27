# Arena

Cooperative match runtime. [Authoritative manifest](../../plugins/arena/plugin.yaml).

## Developer

Start with roomId and distinct player IDs; send playerId, monotonically increasing sequence and optional payload. Read the snapshot and finish with winner IDs and a parseable completion timestamp. The host supplies durable state and rejects forged snapshots.

## Administrator

Enable Arena and grant the four declared operations at the intended room scope. Use its Live sessions extension where available; optional Rooms and Matchmaking dependencies are not automatic authorization. Check active state, player membership and audit on failures.

## Example payloads

arena.start: {"roomId":"r1","players":["a","b"]}; arena.input: {"roomId":"r1","playerId":"a","sequence":1,"payload":{}}; arena.snapshot: {"roomId":"r1"}; arena.result: {"roomId":"r1","winnerIds":["a"],"completedAt":"2026-01-01T12:00:00Z"}.

Payload snippets are arguments to the declared operation, not standalone HTTP routes. Use the [HTTP invocation contract](../http-api.md) and exact manifest version/capability. Host-owned context, internal snapshots and grants must not be supplied by public callers.

## Limitations

Operator-invoked routes do not constitute an end-user realtime game server. Network membership binding must be implemented by the trusted host gateway.
