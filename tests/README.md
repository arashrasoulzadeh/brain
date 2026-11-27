# Test layout

All project tests live under this directory.

```text
tests/
├── app/
├── config/
├── db/
├── migrate/
├── operator/
├── plugin/
├── project/
└── security/
```

Keep unit, contract, API, admin, plugin, integration, and end-to-end tests here, grouped by area. Implementation packages under `internal/` and feature packages under `plugins/` must not contain `*_test.go` files.

Run all tests with:

```bash
GOCACHE=/tmp/brain-go-cache go test ./...
```

## Coverage and verification

Run `npm ci && npm run test:ts` for the TypeScript tests. The coverage workflow measures the Go host and TypeScript surfaces, uploads reports and checks the 100% targets. See the [quality guide](../docs/quality.md) for the scope, limitations and documentation build.
