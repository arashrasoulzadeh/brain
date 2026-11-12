import { dataRequest, projectSchemaName } from "../../plugins/tables/index.js";
import { createRouteHandler } from "../../plugins/sdk-typescript/src/index.js";

const request = dataRequest("dev", "tables", "0.1.0", "tables.read", "users", { filter: { active: true } });
if (request.params.pluginId !== "tables" || request.params.operation !== "tables.read") throw new Error("tables plugin contract failed");
if (typeof createRouteHandler([{ operation: "tables.read", function: "read", capability: "tables.read" }])(request) !== "object") throw new Error("tables handler contract failed");
if (projectSchemaName("dev") !== "project_dev" || projectSchemaName("prod") === projectSchemaName("dev")) throw new Error("project schema isolation failed");
