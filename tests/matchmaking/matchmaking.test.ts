import { matchmakingRequest } from "../../plugins/matchmaking/index.js";
import { createRouteHandler } from "../../plugins/sdk-typescript/src/index.js";

const request = matchmakingRequest("dev", "matchmaking", "0.1.0", "matchmaking.queue", "casual", { playerId: "user-1" });
if (request.params.pluginId !== "matchmaking" || request.params.operation !== "matchmaking.queue") throw new Error("matchmaking plugin contract failed");
if (typeof createRouteHandler([{ operation: "matchmaking.queue", function: "queue", capability: "matchmaking.queue" }])(request) !== "object") throw new Error("matchmaking handler contract failed");
