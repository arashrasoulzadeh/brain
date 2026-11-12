# Phase options and decisions

Each phase has an explicit default path and alternatives to review before implementation expands.

| Phase | Default | Options to consider | Decision point |
|---|---|---|---|
| 00 Foundation | Go host, PostgreSQL integration, SQLite-style local settings | SQLite for fast tests; managed PostgreSQL for deployment; embedded migration runner or external migration tool | Before the first real schema migration |
| 01 Projects/plugins | Host-owned broker and worker contracts | In-process workers for development; subprocess workers for isolation; external services for independently deployed plugins | Before worker supervision |
| 02 Arena | WebSocket rooms with FIFO matchmaking | Dedicated room process; shared room pool; managed realtime provider | Before load testing |
| 03 App data/auth | PostgreSQL row policies and host authorization | Hosted identity provider; self-hosted identity; object storage provider | Before production auth/storage |
| 04 Functions/turns | Go to WASM with wazero | Separate build service; remote build queue; restricted language subset | Before publishing user code |
| 05 UI/workflows | Host-rendered UI plus restricted frames | Declarative-only UI; custom UI bundles; separate admin frontend | Before extension API freeze |
| 06 Clients/release | Browser, Flutter, Android, manual Compose release | Managed VM; Kubernetes; hosted CI/CD deployment | Before production launch |

For every option, record security impact, operational cost, migration cost, test strategy, and rollback plan before changing the default.
