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

Docker is not required for ordinary local development. Run:

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
