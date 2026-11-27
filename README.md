# Brain\n\n**Breaking API namespace:** All HTTP resources require a project prefix (`/{project}/...`) or the reserved `/_system` prefix. See [HTTP API migration](docs/http-api.md). mBaaS

Go multi-tenant backend for apps, websites, and games.

## Status

The host, plugin runtime, and administration surface are implemented and covered by automated checks. See the roadmap for the remaining release and client work. The codebase includes:

- Go HTTP host with `/_system/livez`.
- Configuration validation and isolated local test settings.
- Injectable database owner and bounded readiness retry.
- Migration runner with idempotency checks.
- Runtime DDL protection policy.
- Project resolution with active/inactive isolation.
- Scoped plugin registration, installation, grants, and authorization.
- Go formatting, unit tests, operator race checks, TypeScript lint and tests, admin type checks, and Docker integration checks in GitHub Actions.

See the [roadmap](roadmap.md), [developer guide](docs/developers.md), and [admin guide](docs/admins.md).

## Local development

With Go 1.26.11, Python 3, Node.js 20.15+, npm and Make installed, install the locked Node dependencies once:

```bash
make setup
```

Start the complete stack in a single terminal:

```bash
make dev
```

This launches Brain, the Next.js admin and four independent local Neurons, each with one execution slot. The Neurons authenticate via loopback mutual-TLS Synapse gRPC and must pass the local OS-sandbox probe before connecting. Ctrl+C shuts down the managed stack.

```bash
make dev WORKERS=6 WORKER_SLOTS=2
make dev WORKERS=0
```

Open [the Neurons console](http://127.0.0.1:3000/_system/ui/workers) after signing in. The console reports the real scheduler snapshot, including connections, capacity, runtime profiles, heartbeats and current queue state. The standalone TypeScript pool is separate from existing plugin execution; application routing migration remains outstanding.

For separate server/admin processes without the managed pool, use `make dev server` and `make dev admin`.

The Go server listens on `http://127.0.0.1:18082` and uses local SQLite storage at `.local/brain.sqlite`. The Next.js admin listens on `http://127.0.0.1:3000` and proxies to the server. Stop the stack or individually launched services with Ctrl+C. The local startup script supports `HTTP_ADDRESS`, `PROJECTS_FILE`, `OPERATOR_USERNAME` and `OPERATOR_SECRET` overrides. The admin supports `ADMIN_HOST`, `ADMIN_PORT` and `BRAIN_API_ORIGIN`, for example:

```bash
make dev server HTTP_ADDRESS=127.0.0.1:18090
make dev admin ADMIN_PORT=3100 BRAIN_API_ORIGIN=http://127.0.0.1:18090
```

Development requires no Docker on macOS or Linux. On macOS, `make dev server` selects the built-in Seatbelt (`/usr/bin/sandbox-exec`) worker adapter, with loopback-only HTTP, an allowlist of readable runtime/plugin paths, denied network access, and bounded heap/invocation time. This is a local-development adapter for trusted checkout code, not Linux-equivalent production isolation; Seatbelt's command-line interface is deprecated, and macOS does not enforce the same address-space or process limits. Startup fails closed if the adapter cannot run. On Linux (including WSL2), install bubblewrap and enable user namespaces; the Linux production sandbox is unchanged. Do not use the local script's default credentials in production.

Run the unit tests with:

```bash
GOCACHE=/tmp/brain-go-cache go test ./...
```

PostgreSQL and Compose checks are explicit integration checks and are not run by default.

## Documentation

- [Documentation site](docs/index.md)
- [Developers](docs/developers.md)
- [Administrators](docs/admins.md)
- [Phase options and decisions](docs/phase-options.md)
- [Changelog](CHANGELOG.md)
