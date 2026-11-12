import { arenaRequest, handle, manifest } from "../../plugins/arena/src/index.js";

const request = arenaRequest("dev", "arena.start", "room-1", { players: ["user-1", "user-2"] });
if (manifest.id !== "arena" || request.params.operation !== "arena.start") throw new Error("arena plugin contract failed");
if (typeof handle(request) !== "object") throw new Error("arena handler contract failed");
