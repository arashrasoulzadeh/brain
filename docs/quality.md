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
| Plugin guides | Developer and administrator notes for every manifest | 100% |
| Narrative docs | Review all public interfaces, examples and operating procedures | 100% audit target |

The CI coverage workflow publishes raw profiles and HTML reports even when the target is missed. Do not exclude code merely to reach the target. Go's standard profile measures statements, not branches.

Go CLI integration and packaged runtime smoke tests use an instrumented executable and merge source blocks with unit-test coverage. The host shuts down gracefully on SIGTERM, allowing Go coverage counters to flush. Admin component rendering, transport and DOM interaction tests are included in the TypeScript profile. The separate browser verification workflow runs Chromium against the generated documentation and the real Next.js admin UI with a controlled API fixture; it publishes screenshots and logs. Browser bundle coverage and shell script coverage remain separate unfinished measurement areas. Mobile coverage gates must pass on each revision before their targets are considered achieved. The existing Compose, sandbox, race, lint and type checks remain required. Do not claim repository-wide 100% coverage from partial coverage reports.

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

The generated site uses existing Markdown guides and the host's own manifest documentation command. Do not edit the generated directories. Update the original Markdown or manifest and rebuild. The Pages workflow builds PR previews and deploys only from main.
