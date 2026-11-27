# Analytics

Scoped event capture. [Authoritative manifest](../../plugins/analytics/plugin.yaml).

## Developer

Send a name and bounded scalar fields with analytics.capture. Observe rooms.connected using the declared subscription and trusted event envelope; use event IDs to handle possible redelivery. Do not include credentials or arbitrary records.

## Administrator

Grant analytics.capture and the declared observe.rooms.connected capability when needed. Check observation delivery audit and project state. The admin Analytics operations form captures validated events.

## Example payloads

analytics.capture: {"name":"game.open","fields":{"count":1,"ready":true}}. The room publisher supplies roomId and reconnected; never forge publisher identity.

Payload snippets are arguments to the declared operation, not standalone HTTP routes. Use the [HTTP invocation contract](../http-api.md) and exact manifest version/capability. Host-owned context, internal snapshots and grants must not be supplied by public callers.

## Limitations

The current SQLite event store is not a production analytics warehouse, general query engine or distributed stream. Pure event exporters are not external delivery guarantees.
