import { handle, manifest, roomRequest } from "../../plugins/rooms/src/index.js";

const request = roomRequest("dev", "rooms.create", "room-1", { maxPlayers: 2 });
if (manifest.id !== "rooms" || request.params.operation !== "rooms.create") throw new Error("rooms plugin contract failed");
if (typeof handle(request) !== "object") throw new Error("rooms handler contract failed");
