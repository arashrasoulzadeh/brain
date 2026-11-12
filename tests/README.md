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
