import { IdempotentCreateStore, SchemaPublicationAudit, TableExposureRegistry, VersionedRecord, dataRequest, paginate, planSchemaChange, previewSchemaSQL, projectRoleName, projectSchemaName, projectRow, requirePrimaryKey, requireRecentSchemaAuth, RowPolicy, validateColumnTypes, validateFilter, validateIdentifier, validateValue } from "../../plugins/tables/index.js";
import { createRouteHandler } from "../../plugins/sdk-typescript/src/index.js";

const request = dataRequest("dev", "tables", "0.1.0", "tables.read", "users", { filter: { active: true } });
if (request.params.pluginId !== "tables" || request.params.operation !== "tables.read") throw new Error("tables plugin contract failed");
if (typeof createRouteHandler([{ operation: "tables.read", function: "read", capability: "tables.read" }])(request) !== "object") throw new Error("tables handler contract failed");
if (projectSchemaName("dev") !== "project_dev" || projectSchemaName("prod") === projectSchemaName("dev")) throw new Error("project schema isolation failed");
if (projectRoleName("dev") !== "runtime_project_dev") throw new Error("project role isolation failed");
const policy = new RowPolicy();
if (policy.allows("dev", "user-1") || !policy.grant("dev", "user-1") || !policy.allows("dev", "user-1") || policy.allows("dev", "prod-user")) throw new Error("default deny row policy failed");
const exposures = new TableExposureRegistry();
exposures.expose("dev", "users", ["id", "name"]);
if (!exposures.get("dev", "users")?.columns.includes("id") || exposures.get("prod", "users")) throw new Error("table exposure isolation failed");
requirePrimaryKey(["id", "name"], "id");
try { requirePrimaryKey(["name"], "id"); throw new Error("unsupported primary key accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "supported primary key required") throw error; }
validateColumnTypes({ id: "string", active: "boolean" });
try { validateColumnTypes({ profile: "jsonb" }); throw new Error("unsupported column type accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "unsupported column type") throw error; }
if (Object.keys(projectRow({ id: "1", secret: "x" }, ["id"])).length !== 1) throw new Error("unapproved column exposed");
validateFilter({ active: { eq: true } }, ["active"]);
try { validateFilter({ secret: { eq: true } }, ["active"]); throw new Error("unapproved filter accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "unapproved filter") throw error; }
const page = paginate([1, 2, 3], undefined, 2);
if (page.items.join(",") !== "1,2" || !page.nextCursor || paginate([1, 2, 3], page.nextCursor, 2).items[0] !== 3) throw new Error("cursor pagination failed");
validateIdentifier("users"); validateValue("name", "alice");
try { validateIdentifier("users;drop"); throw new Error("unsafe identifier accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "invalid identifier") throw error; }
const record = new VersionedRecord({ id: "1", name: "alice" });
record.update(0, { id: "1", name: "bob" });
try { record.update(0, { id: "1", name: "carol" }); throw new Error("stale write accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "stale record version") throw error; }
const creates = new IdempotentCreateStore<{ id: string }>();
let calls = 0;
const created = creates.create("dev", "key-1", () => { calls += 1; return { id: "row-1" }; });
if (creates.create("dev", "key-1", () => { calls += 1; return { id: "row-2" }; }).id !== created.id || calls !== 1) throw new Error("idempotent create failed");
const schemaPlan = planSchemaChange("dev", "users", { kind: "add_column", name: "display_name", type: "string" });
if (schemaPlan.projectId !== "dev" || schemaPlan.table !== "users" || schemaPlan.change.kind !== "add_column") throw new Error("guided schema plan failed");
try { planSchemaChange("dev", "users", { kind: "drop_table", name: "users" } as never); throw new Error("unsafe schema change accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "unsupported schema change") throw error; }
if (previewSchemaSQL(schemaPlan) !== 'ALTER TABLE "project_dev"."users" ADD COLUMN "display_name" TEXT;') throw new Error("schema SQL preview failed");
try { previewSchemaSQL({ ...schemaPlan, table: "users;drop" }); throw new Error("unsafe SQL preview accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "invalid schema target") throw error; }
requireRecentSchemaAuth(9_000, 10_000);
try { requireRecentSchemaAuth(1_000, 301_001); throw new Error("stale schema auth accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "recent authentication required") throw error; }
const audit = new SchemaPublicationAudit();
audit.record("dev", "operator-1", schemaPlan, 123);
if (audit.list("dev").length !== 1 || audit.list("prod").length !== 0 || audit.list("dev")[0].actorId !== "operator-1") throw new Error("schema publication audit failed");
