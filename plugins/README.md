# Plugin catalog

The authoritative inventory is the set of `plugins/*/plugin.yaml` manifests. Each executable plugin declares its entrypoint, source files, runtime imports, capabilities and routes. A source-free manifest can use approved declarative helpers instead. The documentation site generates the reference directly from validated manifests, and maintains a developer/administrator guide for every manifest.

| Installed plugin ID | Purpose |
| --- | --- |
| access-management | Durable project-scoped invocation rules |
| analytics | Validated events and room observations |
| arena | Cooperative match state, actions and results |
| auth | Project accounts, sessions and redacted operator inventory |
| dashboard | Trusted host-only operator administration |
| files | Scoped object content and metadata |
| functions | Function route contracts and authoring controls |
| games | Game-specific rules, including Tic-Tac-Toe |
| invites | Room invitation allocation, acceptance and cancellation |
| matchmaking | Queue, ticket status and assignment |
| pvp | Generic turn, revision, replay and private-view primitives |
| rooms | Game-independent room lifecycle and membership |
| subscriptions | Bounded room-connection observation polling |
| tables | Current SQLite-backed table/row compatibility API |
| template | Example plugin and model/pipeline patterns |

The source-free [catalog example](../examples/declarative-plugin/plugin.yaml) is not discovered from the normal plugin root. Go/WASM, distributed workflows and Eye export have supporting libraries or roadmap material, not separate registered plugins in this inventory.

See [developer guide](../docs/developers.md), [SDK reference](../docs/sdk-reference.md) and [admin guide](../docs/admins.md). Product-specific Go storage adapters remain transitional; do not add new business logic to them.
