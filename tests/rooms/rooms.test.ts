import { createHandler, roomRequest } from "../../plugins/rooms/index.js";

const request = roomRequest("dev", "rooms", "0.1.0", "rooms.create", "room-1", { maxPlayers: 2 });
if (request.params.pluginId !== "rooms" || request.params.operation !== "rooms.create") throw new Error("rooms plugin contract failed");
if (typeof createHandler([{ operation: "rooms.create", function: "create", capability: "rooms.create" }])(request) !== "object") throw new Error("rooms handler contract failed");
