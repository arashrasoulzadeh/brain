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

The host gateway supports authenticated player input, not a broadcast simulation protocol. Start an arena with the exact project-qualified user IDs (for example `alpha:alice` and `alpha:bob`). Grant each user `arena.connect` (`arena.snapshot` capability) and `arena.player.input` (`arena.input` capability) for that room. Obtain a ticket by POSTing `{"path":"/arena/r1/input"}` to `/alpha/ws`, then connect with the `brain.v1` and `ticket.<token>` subprotocols. Send `{"sequence":1,"payload":{}}`. The worker checks active membership and sequence on every input; the host derives the player from the verified session and persists the sequence. Reconnect requires a fresh ticket. Broadcasts, automatic matchmaking and high-frequency simulation are separate features.
