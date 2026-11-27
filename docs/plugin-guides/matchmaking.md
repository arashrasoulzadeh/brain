# Matchmaking

Queue and ticket assignment. [Authoritative manifest](../../plugins/matchmaking/plugin.yaml).

## Developer

Use id as the actual ticket resource. Queue creates a ticket, assign and cancel change its persisted status, and status reads the host-owned snapshot. Never treat ticket metadata supplied by the caller as durable state.

## Administrator

Install and enable with the declared grants. Rooms is optional. The Matchmaking tab in Live sessions supports inventory, status, queue, assignment and cancellation. Restrict mutation capabilities to operators who need them.

## Example payloads

matchmaking.queue / matchmaking.status / matchmaking.assign / matchmaking.cancel: {"id":"ticket-1"} (create before changing status).

Payload snippets are arguments to the declared operation, not standalone HTTP routes. Use the [HTTP invocation contract](../http-api.md) and exact manifest version/capability. Host-owned context, internal snapshots and grants must not be supplied by public callers.

## Limitations

This is a basic ticket state adapter, not a distributed matchmaking scheduler or automatic room allocation service.
