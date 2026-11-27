# Analytics

Scoped event capture. [Authoritative manifest](../../plugins/analytics/plugin.yaml).

## Developer

Send a name and bounded scalar fields with analytics.capture. Observe rooms.connected using the declared subscription and trusted event envelope; use event IDs to handle possible redelivery. Do not include credentials or arbitrary records.

## Administrator

Grant analytics.capture and the declared observe.rooms.connected capability when needed. Grant analytics.read for exact-name history and analytics.summary separately for project-wide counts. Existing installations must explicitly grant the new capabilities and satisfy any Access Management rules before the updated admin workspace is eligible. The admin Analytics operations form captures validated events, reads recent matching events and shows event-count aggregates.

## Example payloads

analytics.capture: {"name":"game.open","fields":{"count":1,"ready":true}}. The room publisher supplies roomId and reconnected; never forge publisher identity.

Payload snippets are arguments to the declared operation, not standalone HTTP routes. Use the [HTTP invocation contract](../http-api.md) and exact manifest version/capability. Host-owned context, internal snapshots and grants must not be supplied by public callers.

## Limitations

The current SQLite event store is not a production analytics warehouse, general query engine or distributed stream. Pure event exporters are not external delivery guarantees.

History accepts name, optional offset (default 0), and limit (1–20; default 20). Summary accepts offset and limit (1–100; default 20), sorts event names alphabetically and returns counts rather than event fields. Both responses include totals and next_offset when more results are available. History returns newest appended events first, bounded by the worker payload ceiling, with timestamps on newly captured events. Older persisted events do not have a reconstructed timestamp. Pagination is over the current event store, not a consistent snapshot; concurrent capture can shift offsets. Reads use the existing host-owned SQLite event state and never caller-supplied snapshots. This implementation still reads the stored event blob and is not suitable for unbounded warehouse-scale workloads.

History and event counts accept optional paired `from` and `until` RFC3339 timestamps, using an inclusive start and exclusive end (maximum 366 days). The project-wide `analytics.timeline` operation uses the existing analytics.summary capability, requires both timestamps and bucket `hour` or `day`, and optionally filters by event name. Up to 168 UTC buckets are returned, including zero-count buckets. Legacy undated events are reported separately and excluded from time-filtered counts. This remains a bounded read over the existing event snapshot, not a warehouse or streaming query engine.
