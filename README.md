# Brain mBaaS

Go multi-tenant backend for apps, websites, and games.

## Status

The host, plugin runtime, and administration surface are implemented and covered by automated checks. See the roadmap for the remaining release and client work. The codebase includes:

- Go HTTP host with `/livez`.
- Configuration validation and isolated local test settings.
- Injectable database owner and bounded readiness retry.
- Migration runner with idempotency checks.
- Runtime DDL protection policy.
- Project resolution with active/inactive isolation.
- Scoped plugin registration, installation, grants, and authorization.
- Go formatting, unit tests, operator race checks, TypeScript lint and tests, admin type checks, and Docker integration checks in GitHub Actions.

See the [roadmap](roadmap.md), [developer guide](docs/developers.md), and [admin guide](docs/admins.md).

## Local development

With Go, Node.js 20+, npm, GNU Make and Linux bubblewrap available, install the locked Node dependencies once:

```bash
make setup
```

Start the server and admin in separate terminals:

```bash
make dev server
make dev admin
```

The Go server listens on `http://127.0.0.1:18082` and uses local SQLite storage at `.local/brain.sqlite`. The Next.js admin listens on `http://127.0.0.1:3000` and proxies to the server. Stop either process with Ctrl+C. The local startup script supports `HTTP_ADDRESS`, `PROJECTS_FILE`, `OPERATOR_USERNAME` and `OPERATOR_SECRET` overrides. The admin supports `ADMIN_HOST`, `ADMIN_PORT` and `BRAIN_API_ORIGIN`, for example:

```bash
make dev server HTTP_ADDRESS=127.0.0.1:18090
make dev admin ADMIN_PORT=3100 BRAIN_API_ORIGIN=http://127.0.0.1:18090
```

Development requires no Docker and keeps the production worker sandbox enabled. Use a Linux environment with permitted user namespaces (including an appropriately configured WSL2 environment). Do not use the local script's default credentials in production.

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
