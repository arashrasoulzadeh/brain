import { arenaRequest } from "../../plugins/arena/index.js";
import { createRouteHandler } from "../../plugins/sdk-typescript/src/index.js";

const request = arenaRequest("dev", "arena", "0.1.0", "arena.start", "room-1", { players: ["user-1", "user-2"] });
if (request.params.pluginId !== "arena" || request.params.operation !== "arena.start") throw new Error("arena plugin contract failed");
if (typeof createRouteHandler([{ operation: "arena.start", function: "start", capability: "arena.start" }])(request) !== "object") throw new Error("arena handler contract failed");
