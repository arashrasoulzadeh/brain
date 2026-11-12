# Brain mBaaS

Go multi-tenant backend for apps, websites, and games.

## Status

Phase 00 foundation is in progress. The current verified code includes:

- Go HTTP host with `/livez`.
- Configuration validation and isolated local test settings.
- Injectable database owner and bounded readiness retry.
- Migration runner with idempotency checks.
- Runtime DDL protection policy.
- Project resolution with active/inactive isolation.
- Scoped plugin registration, installation, grants, and authorization.
- GitHub Actions formatting and Go test checks.

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
