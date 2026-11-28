# Neurons

Routes: `/_system/ui/workers` (cluster console) and `/{project}/ui/workers` (compatibility link).

The system-level console is available to operators without an active project or plugin. It reads `GET /_system/admin/neurons`, displays connected identities, runtime profiles, occupied/free slots, last heartbeat, running assignments and queue depth from the bound Cortex scheduler. An unconfigured scheduler or failed poll displays unavailable metrics, not invented zeroes. The list refreshes every five seconds when visible and can be refreshed manually.

The existing project-specific dedicated plugin worker status and explicit startup recovery remain under Legacy plugin workers. Generic Neurons and legacy workers are not combined. A native development session launched with `make dev` registers four independently running Neurons via mutually authenticated Synapse gRPC by default; without a bound scheduler the server explicitly reports that the pool is not configured. The new standalone TypeScript runtime is distinct from the legacy project plugin workers. CPU/memory, execution history, completed/failed counters, latency, drain and managed restart actions are not yet provided by the host.

Do not disable sandbox isolation or broaden permissions as a workaround for a failed worker. See [HTTP API](../http-api.md).
