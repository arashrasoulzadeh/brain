# Database

Route: `/{project}/ui/database`

Runs project-scoped table operations through the Tables plugin. Requires Tables to be ready and enabled. Uses the shared admin sidebar.

This is the Tables manifest alias, not an SQL console. Select an authorized table name from inventory, query its inferred schema or read a row page (limit 1–100, offset default 0). Read results show zero-based row indexes, a table snapshot revision, and the next offset when more rows exist. Append a JSON row, replace a row, or delete a row using the explicitly confirmed operation. For replacement or deletion, copy the index and revision from the current row page. If the table changed, the host rejects the stale revision; read a fresh page before retrying. Existing installations must explicitly grant tables.update and tables.delete and satisfy applicable access rules. The current adapter persists SQLite row snapshots. Do not enter arbitrary SQL or assume the displayed rows are a PostgreSQL schema.

See the [admin route index](README.md) and [HTTP API](../http-api.md).

Use Profile observed column types to inspect observed JSON types, missing/present counts and the revision. A separate `tables.admin.schema` capability authorizes creating an empty table and confirmed bulk add/rename/drop of JSON fields. Bulk changes require the current revision, reject collisions and commit atomically; they are limited to 10,000 rows and 4 MiB. Adding a column affects existing rows only (default null), not future write constraints. This is not SQL DDL or a relational schema enforcement system.
