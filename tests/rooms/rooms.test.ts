import { projectRoomID, roomRequest } from "../../plugins/rooms/index.js";
import { createRouteHandler } from "../../plugins/sdk-typescript/src/index.js";

const request = roomRequest("dev", "rooms", "0.1.0", "rooms.create", "room-1", { maxPlayers: 2 });
if (request.params.pluginId !== "rooms" || request.params.operation !== "rooms.create") throw new Error("rooms plugin contract failed");
if (typeof createRouteHandler([{ operation: "rooms.create", function: "create", capability: "rooms.create" }])(request) !== "object") throw new Error("rooms handler contract failed");
if (projectRoomID("dev", "room-1") !== "dev:room-1") throw new Error("project room id was not scoped");
try { projectRoomID("dev", "prod:room"); throw new Error("invalid room id accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "invalid project room id") throw error; }
