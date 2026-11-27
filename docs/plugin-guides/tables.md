# Tables

SQLite row and schema compatibility API. [Authoritative manifest](../../plugins/tables/plugin.yaml).

## Developer

Use table as the resource. Write a JSON row object, read stored rows, and derive schema from stored row keys. The host strips caller-supplied snapshots. Separate schema, read and write grants.

## Administrator

Use the Database page and table inventory. Confirm writes and grant the minimum resource scope. Use tables.read with offset and limit (1–100) for bounded pages, then tables.update or tables.delete with table, zero-based row_index and the returned revision; update also requires a replacement JSON row. These new mutation capabilities require explicit grants. The host checks the snapshot and uses a compare-and-swap on the stored JSON in the commit transaction. Indexes are not stable row IDs, and any change to the table invalidates the revision. The legacy tables.read without pagination arguments and tables.write append contract remain compatible. This is snapshot CRUD, not arbitrary SQL or a full relational schema interface.

## Example payloads

tables.write: {"table":"messages","row":{"id":"one","text":"hello"}}; tables.read / tables.schema: {"table":"messages"}.

Payload snippets are arguments to the declared operation, not standalone HTTP routes. Use the [HTTP invocation contract](../http-api.md) and exact manifest version/capability. Host-owned context, internal snapshots and grants must not be supplied by public callers.

## Limitations

PostgreSQL migrations, full relational schema publication and row-level access controls are separate primitives/release work, not the mounted SQLite route API.

The host supports `tables.schema.profile` (existing schema grant) and `tables.admin.create`, `tables.admin.add-column`, `tables.admin.rename-column`, `tables.admin.drop-column` (new exact-table `tables.admin.schema` grant). Profile returns a paged observed column summary and snapshot revision; bulk changes require that revision and reject target collisions or stale state. They affect stored JSON records, not relational constraints, indexes or future inserts. No arbitrary SQL is executed.
