# Neurons

Routes: `/_system/ui/workers` (cluster console) and `/{project}/ui/workers` (compatibility link).

The system-level console is available to operators without an active project or plugin. It reads `GET /_system/admin/neurons`, displays connected identities, runtime profiles, occupied/free slots, last heartbeat, running assignments and queue depth from the bound Cortex scheduler. An unconfigured scheduler or failed poll displays unavailable metrics, not invented zeroes. The list refreshes every five seconds when visible and can be refreshed manually.

The existing project-specific dedicated plugin worker status and explicit startup recovery remain under Legacy plugin workers. Generic Neurons and legacy workers are not combined. A native development session launched with `make dev` registers four independently running Neurons via mutually authenticated Synapse gRPC by default; without a bound scheduler the server explicitly reports that the pool is not configured. The new standalone TypeScript runtime is distinct from the legacy project plugin workers. The host provides recent execution history, outcome counters, latency, execution CPU time and peak resident memory, plus drain and managed restart actions. Per-Neuron performance aggregates the last 128 cluster-wide retained outcomes, including historical identities. Missing resource measurements are excluded from means; sample counts are displayed. These measurements do not represent instantaneous CPU utilization or live resident memory.

Do not disable sandbox isolation or broaden permissions as a workaround for a failed worker. See [HTTP API](../http-api.md).

## Prometheus

Set a dedicated random scrape token before starting Brain. The endpoint is disabled when the token is unset and never accepts operator session credentials in place of its configured token. Use a protected token file for Prometheus and TLS for non-loopback connections.

```sh
export BRAIN_METRICS_TOKEN="$(openssl rand -hex 32)"
make dev WORKERS=10
```

Scrape Brain directly on port 18082, not the Next.js server. Example Prometheus configuration (provide the matching token in the protected file):

```yaml
scrape_configs:
  - job_name: brain-cortex
    metrics_path: /_system/metrics
    authorization:
      type: Bearer
      credentials_file: /run/secrets/brain_metrics_token
    static_configs:
      - targets: ["127.0.0.1:18082"]
```

The exporter uses the `brain_cortex_` prefix. Process-lifetime assignment and terminal-outcome metrics are counters. Capacity, queue depth, recent execution counts, mean/p95 duration, mean wait, CPU seconds and peak RSS bytes are gauges. Recent resource gauges are omitted without a measurement; corresponding sample-count gauges remain available. No execution IDs, trace IDs, project IDs, inputs, outputs or credentials are metric labels. Neuron identity is the only worker-specific label; operate a bounded set of stable enrolled identities.
