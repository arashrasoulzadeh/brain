import { createHandler, dataRequest } from "../../plugins/tables/index.js";

const request = dataRequest("dev", "tables", "0.1.0", "tables.read", "users", { filter: { active: true } });
if (request.params.pluginId !== "tables" || request.params.operation !== "tables.read") throw new Error("tables plugin contract failed");
if (typeof createHandler([{ operation: "tables.read", function: "read", capability: "tables.read" }])(request) !== "object") throw new Error("tables handler contract failed");
