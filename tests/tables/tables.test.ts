import { TableExposureRegistry, dataRequest, projectRoleName, projectSchemaName, requirePrimaryKey, RowPolicy } from "../../plugins/tables/index.js";
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
