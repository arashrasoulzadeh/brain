import { createArena, MatchResultReader, MatchResultSaver, arenaStartCommand } from "../../plugins/arena/index.js";
import { ConnectionTicketIssuer, ConnectionTicketConsumer, RoomDirectory, authorizeRoomCommand } from "../../plugins/rooms/index.js";
import { MatchmakingQueue } from "../../plugins/matchmaking/index.js";

const browserOne = createArena("dev", "gate-room");
const browserTwo = createArena("dev", "gate-room");
arenaStartCommand(browserOne, "dev", "gate-room", ["user-1", "user-2"]);
arenaStartCommand(browserTwo, "dev", "gate-room", ["user-1", "user-2"]);
if (browserOne.getState() !== "active" || browserTwo.getState() !== "active") throw new Error("two-browser match did not start");
const rooms = new RoomDirectory();
rooms.create("dev", "gate-room", 2);
rooms.join("dev", "gate-room", "user-1", 0);
const ticketMap = new Map(); const ticketIssuer = new ConnectionTicketIssuer(ticketMap); const ticketConsumer = new ConnectionTicketConsumer(ticketMap);
const reconnectTicket = ticketIssuer.issue("dev", "gate-room", "user-1", 1000, 60);
if (!rooms.reconnect("dev", "gate-room", "user-1") || !ticketConsumer.consume(reconnectTicket.token, "dev", "gate-room", "user-1", 1001)) throw new Error("dropped connection did not recover");
try { authorizeRoomCommand(rooms.get("dev", "gate-room"), "dev", "user-9", "arena.input"); throw new Error("denied plugin action accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "room membership required") throw error; }
const completed = browserOne.result(["user-1"], "2026-11-13T00:00:00.000Z");
const resultMap = new Map();
new MatchResultSaver(resultMap).save(completed);
if (new MatchResultReader(resultMap).get(completed.resultId)?.projectId !== "dev") throw new Error("durable result was not persisted");
try { rooms.get("prod", "gate-room"); throw new Error("private room crossed project boundary"); } catch (error) { if (!(error instanceof Error) || error.message !== "room not found") throw error; }
if (rooms.isMember("dev", "gate-room", "user-1") !== true || rooms.isMember("dev", "gate-room", "user-9") !== false) throw new Error("private room membership gate failed");
const queue = new MatchmakingQueue();
const first = queue.enqueue("dev", "ranked", "user-1", 1000);
queue.enqueue("dev", "ranked", "user-2", 1001);
const decision = queue.assign("dev", "ranked", 2, 1002);
if (!decision || decision.ticketIds[0] !== first.id) throw new Error("FIFO matchmaking gate failed");
const eightPlayers = Array.from({ length: 8 }, (_, index) => `player-${index + 1}`);
const eightArena = createArena("dev", "eight-player-gate");
eightArena.start(eightPlayers);
if (eightArena.players.length !== 8) throw new Error("eight-player gate result missing");
