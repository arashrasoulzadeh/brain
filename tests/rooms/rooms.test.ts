import { RoomDirectory, canTransitionRoom, projectRoomID, roomRequest, validateMembership } from "../../plugins/rooms/index.js";
import { createRouteHandler } from "../../plugins/sdk-typescript/src/index.js";

const request = roomRequest("dev", "rooms", "0.1.0", "rooms.create", "room-1", { maxPlayers: 2 });
if (request.params.pluginId !== "rooms" || request.params.operation !== "rooms.create") throw new Error("rooms plugin contract failed");
if (typeof createRouteHandler([{ operation: "rooms.create", function: "create", capability: "rooms.create" }])(request) !== "object") throw new Error("rooms handler contract failed");
if (projectRoomID("dev", "room-1") !== "dev:room-1") throw new Error("project room id was not scoped");
try { projectRoomID("dev", "prod:room"); throw new Error("invalid room id accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "invalid project room id") throw error; }
validateMembership("dev", "room-1", [{ playerId: "user-1", slot: 0 }, { playerId: "user-2", slot: 1 }], 2);
try { validateMembership("dev", "room-1", [{ playerId: "user-1", slot: 0 }, { playerId: "user-2", slot: 0 }], 2); throw new Error("duplicate player slot accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "invalid player slots") throw error; }
if (!canTransitionRoom("waiting", "matching") || !canTransitionRoom("matching", "active") || canTransitionRoom("closed", "active")) throw new Error("room lifecycle transition invalid");
const rooms = new RoomDirectory();
if (rooms.create("dev", "room-new", 4).state !== "waiting") throw new Error("room did not start waiting");
try { rooms.create("dev", "room-new", 4); throw new Error("duplicate room accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "room already exists") throw error; }
try { rooms.get("prod", "room-new"); throw new Error("cross-project room lookup accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "room not found") throw error; }
rooms.join("dev", "room-new", "user-1", 0);
try { rooms.join("prod", "room-new", "user-2", 1); throw new Error("cross-project join accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "room not found") throw error; }
try { rooms.join("dev", "room-new", "user-1", 1); throw new Error("duplicate membership accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "invalid membership") throw error; }
