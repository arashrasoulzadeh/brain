# Plugin catalog

All Brain product features live here as TypeScript plugins. The Go host in `internal/` provides shared infrastructure only.

## Required layout

```text
plugins/<plugin-id>/
├── plugin.yaml        # authoritative metadata, routes, functions, permissions
└── index.ts           # serverless TypeScript functions only
```

## Feature plugins

| Plugin | Purpose | Phase |
|---|---|---:|
| `dashboard` | Operator UI and project administration | 01/05 |
| `auth` | End-user identity and sessions | 01/03 |
| `tables` | Project PostgreSQL tables and CRUD | 03 |
| `files` | Private object storage | 03 |
| `subscriptions` | App realtime subscriptions | 03 |
| `functions` | Go/WASM functions | 04 |
| `workflows` | Project and release workflows | 04/05 |
| `rooms` | Room lifecycle and membership | 02 |
| `invites` | Private-room invites | 02 |
| `matchmaking` | Queueing and assignment | 02 |
| `arena` | Co-op game runtime | 02 |
| `turn-rules` | Turn-based game runtime | 04 |
| `analytics` | Validated event contract | 01 |
| `eye-export` | Eye event export | 02 |

## Runtime rule

Every plugin runs as a TypeScript/Node.js worker and communicates through the host RPC contract. User-authored functions use the stricter sandboxed runtime.

## Completion contract

The host supplies the runtime and package dependencies, discovers `plugin.yaml`, validates routes and grants, and generates developer/administrator documentation from the YAML contract. Tests live under the top-level `tests/` directory.
