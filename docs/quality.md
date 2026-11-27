# Quality and coverage

Passing tests and coverage are different measures. A coverage gate does not prove that assertions are meaningful or that a deployment is safe.

## Coverage targets

| Area | Metric | Target |
| --- | --- | --- |
| Go host and command | Executed statements | 100% |
| TypeScript plugins, SDK, JavaScript client, admin and worker | Lines, branches, functions, statements per file | 100% |
| Flutter client | Executed source lines per file | 100% |
| Kotlin client | Lines, branches, instructions and methods per class | 100% |
| Plugin manifests | Generated reference for each manifest | 100% |
| Plugin developer and administrator guides | One guide per discovered manifest and example | 100% |
| HTTP/admin/function commands, admin pages and SDK exports | Exact source inventory audit | 100% |
| Narrative docs | Review all public interfaces, examples and operating procedures | 100% audit target |

The CI coverage workflow publishes raw profiles and HTML reports even when the target is missed. Do not exclude code merely to reach the target. Go's standard profile measures statements, not branches.

Shell scripts, browser end-to-end flows and production integration coverage are separate unfinished measurement areas. Flutter and Kotlin coverage gates are configured but must pass before their targets are considered achieved. The existing Compose, sandbox, race, lint and type checks remain required. Do not claim repository-wide 100% coverage from partial coverage reports.

## Reproduce the site

From the repository root, install the documentation requirements into an isolated Python environment:

```bash
python -m venv .local/docs-venv
source .local/docs-venv/bin/activate
pip install -r docs/requirements.txt
mkdir -p .local
go build -o .local/brain ./cmd/brain
python -m unittest discover -s tests/documentation
python scripts/build-docs.py --contract .local/brain
mkdocs build --strict
mkdocs serve
```

The generated site uses existing Markdown guides and the host's own manifest documentation command in Markdown and TypeScript formats. The documentation inventory test fails for missing or stale HTTP routes, admin commands, function commands, plugin guides or undocumented public SDK exports. The build also requires developer, administrator, examples and limitations sections in every plugin guide. Inventory coverage is not proof that every narrative claim is correct; review examples and operational procedures independently. Do not edit the generated directories. Update the original Markdown or manifest and rebuild. The Pages workflow builds PR previews and deploys only from main.
