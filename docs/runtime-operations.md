# Runtime configuration and operations

## Start the current host

Use Go as pinned in [go.mod](../go.mod), Node 20.15+ with the supported WASM trap-handler flag, npm and Make. Run `make setup`, then `make dev server` and `make dev admin` in separate terminals. Development defaults are loopback server port 18082 and admin port 3000. Never use local credentials or bind local development publicly.

| Setting | Scope | Meaning |
| --- | --- | --- |
| HTTP_ADDRESS | host | Required bind address. Local default 127.0.0.1:18082. |
| OPERATOR_USERNAME, OPERATOR_SECRET | host | Required bootstrap credentials. Local defaults admin/secret only. |
| PROJECTS_FILE | host | Required durable SQLite path (not empty, :memory: or file: URI). |
| RUNTIME_DSN, MIGRATION_DSN | host | PostgreSQL application storage not wired; nonempty values refuse normal startup. |
| BRAIN_LOCAL_MACOS | host | Set by local launcher on macOS; opt-in Seatbelt adapter, restricted to 127.0.0.1. |
| ADMIN_HOST, ADMIN_PORT | Make/admin | Admin bind address and port. |
| BRAIN_API_ORIGIN | admin | Host origin for Next.js rewrites. Default http://127.0.0.1:18082. |

Linux needs Bubblewrap and usable namespaces. macOS uses the deprecated OS Seatbelt launcher for trusted local development, not production-equivalent isolation. Neither adapter may silently fall back to unsandboxed workers. Worker policy denies network access, bounds execution, and exposes only declared readable source/runtime paths. Linux caps address space at 2 GiB, shared UID process/thread count at 512 and worker CPU time at 30 seconds; invocation wall time is 2 seconds. See [host startup](../cmd/brain/main.go), [Linux launcher](../scripts/run-sandboxed-typescript-worker.sh) and [macOS launcher](../scripts/run-macos-typescript-worker.sh).

## Command-line reference

| Command | Purpose |
| --- | --- |
| `go run ./cmd/brain` | Start configured application. |
| `go run ./cmd/brain health-only` | Liveness fixture only; no application routes or readiness. |
| `go run ./cmd/brain plugin-contract plugins/games/plugin.yaml markdown` | Render validated manifest docs. |
| `go run ./cmd/brain plugin-contract plugins/games/plugin.yaml typescript` | Render TypeScript contract. |
| `go run ./cmd/brain bootstrap-operator` | Validate bootstrap credentials in an ephemeral store; not a persistent account provisioner. |
| `go run ./cmd/brain bootstrap-dashboard` | Validate dashboard bootstrap. |
| `go run ./cmd/brain verify-object-manifest path.json` | Validate object-version manifest. |

## Durable data and restart

The current host is single-instance SQLite-based. The project store, plugin lifecycle, plugin model records, product compatibility state and observation outbox share the SQLite database. Every connection uses a 5-second busy timeout (see [shared SQLite configuration](../internal/sqliteconfig/config.go)). This bounds waiting on a concurrent writer; it does not make the store distributed or eliminate all contention. A failed lifecycle persistence closes authorization until state is successfully saved. Check readiness and audit rather than blindly retrying a mutation.

For backup, stop incoming traffic and writes; stop the host or use an approved consistent SQLite backup procedure. Preserve the database and any required WAL/SHM state as a consistent unit, plus the function draft/log files and the exact plugin sources, manifests and runtime release. Verify a restore on an isolated copy, including project isolation, grants, worker readiness, session behavior and pending observations. Do not copy just the live database file while writes are in flight. Protect credential hashes, tokens, drafts and logs. Operator sessions are in memory and need new login after restart. Do not apply the PostgreSQL/object-store restore procedure to this SQLite deployment.

## Diagnose startup

- Missing dependency or TypeScript loader: run `make setup`.
- Missing bwrap on Linux: install Bubblewrap and verify namespace permissions; do not disable the sandbox.
- macOS non-loopback address: use 127.0.0.1 and the local launcher.
- Live but not ready: inspect durable storage and enabled worker health, not just HTTP process status.
- Worker startup or authorization failure: check declared sources/imports, dependency versions, grants, access rules and audit. Do not widen permissions to hide a failing check.

See [deployment runbook](deployment-runbook.md) for release rehearsals and [restore runbook](restore-runbook.md) for the separate PostgreSQL/object-storage recovery design.
