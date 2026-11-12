import { ArenaStateMachine, MatchResultStore, arenaStartCommand } from "../../plugins/arena/index.js";
import { ConnectionTickets, RoomDirectory, authorizeRoomCommand } from "../../plugins/rooms/index.js";

const browserOne = new ArenaStateMachine("dev", "gate-room");
const browserTwo = new ArenaStateMachine("dev", "gate-room");
arenaStartCommand(browserOne, "dev", "gate-room", ["user-1", "user-2"]);
arenaStartCommand(browserTwo, "dev", "gate-room", ["user-1", "user-2"]);
if (browserOne.getState() !== "active" || browserTwo.getState() !== "active") throw new Error("two-browser match did not start");
const rooms = new RoomDirectory();
rooms.create("dev", "gate-room", 2);
rooms.join("dev", "gate-room", "user-1", 0);
const tickets = new ConnectionTickets();
const reconnectTicket = tickets.issue("dev", "gate-room", "user-1", 1000, 60);
if (!rooms.reconnect("dev", "gate-room", "user-1") || !tickets.consume(reconnectTicket.token, "dev", "gate-room", "user-1", 1001)) throw new Error("dropped connection did not recover");
try { authorizeRoomCommand(rooms.get("dev", "gate-room"), "dev", "user-9", "arena.input"); throw new Error("denied plugin action accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "room membership required") throw error; }
const completed = browserOne.result(["user-1"], "2026-11-13T00:00:00.000Z");
const store = new MatchResultStore();
store.save(completed);
if (store.get(completed.resultId)?.projectId !== "dev") throw new Error("durable result was not persisted");
