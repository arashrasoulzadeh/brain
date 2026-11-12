import { dataRequest, handle, manifest } from "../../plugins/tables/src/index.js";

const request = dataRequest("dev", "tables.read", "users", { filter: { active: true } });
if (manifest.id !== "tables" || request.params.operation !== "tables.read") throw new Error("tables plugin contract failed");
if (typeof handle(request) !== "object") throw new Error("tables handler contract failed");
