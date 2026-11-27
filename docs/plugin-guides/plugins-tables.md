# Tables

## Developer

Write with `tables.write` payload `{"table":"messages","row":{"id":"one","text":"hello"}}`. Read rows and inspect columns with `tables.read` and `tables.schema`, using the same table name. The table name is the resource identity. The host performs storage planning and rechecks authorization before committing writes.

## Administrator

Grant read, write and schema independently and use explicit table-level resource rules. Verify SQLite persistence and restore with a dedicated fixture project. Do not treat dynamic table columns as a substitute for an approved schema or row-level authorization; review the model contract for structured state.
