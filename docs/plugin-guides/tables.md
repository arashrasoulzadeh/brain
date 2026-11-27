# Tables

SQLite row and schema compatibility API. [Authoritative manifest](../../plugins/tables/plugin.yaml).

## Developer

Use table as the resource. Write a JSON row object, read stored rows, and derive schema from stored row keys. The host strips caller-supplied snapshots. Separate schema, read and write grants.

## Administrator

Use the Database page and table inventory. Confirm writes and grant the minimum resource scope. Current row writes append rather than exposing arbitrary SQL or a full relational CRUD interface.

## Example payloads

tables.write: {"table":"messages","row":{"id":"one","text":"hello"}}; tables.read / tables.schema: {"table":"messages"}.

Payload snippets are arguments to the declared operation, not standalone HTTP routes. Use the [HTTP invocation contract](../http-api.md) and exact manifest version/capability. Host-owned context, internal snapshots and grants must not be supplied by public callers.

## Limitations

PostgreSQL migrations, full relational schema publication and row-level access controls are separate primitives/release work, not the mounted SQLite route API.
