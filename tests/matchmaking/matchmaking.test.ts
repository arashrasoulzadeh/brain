import { handle, manifest, matchmakingRequest } from "../../plugins/matchmaking/src/index.js";

const request = matchmakingRequest("dev", "matchmaking.queue", "casual", { playerId: "user-1" });
if (manifest.id !== "matchmaking" || request.params.operation !== "matchmaking.queue") throw new Error("matchmaking plugin contract failed");
if (typeof handle(request) !== "object") throw new Error("matchmaking handler contract failed");
