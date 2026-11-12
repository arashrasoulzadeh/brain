import { fileRequest } from "../../plugins/files/index.js";
import { createRouteHandler } from "../../plugins/sdk-typescript/src/index.js";

const request = fileRequest("dev", "files", "0.1.0", "files.upload", "avatar", { contentType: "image/png" });
if (request.params.pluginId !== "files" || request.params.operation !== "files.upload") throw new Error("files plugin contract failed");
if (typeof createRouteHandler([{ operation: "files.upload", function: "upload", capability: "files.upload" }])(request) !== "object") throw new Error("files handler contract failed");
