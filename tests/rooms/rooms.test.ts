import { ConnectionTickets, CommandDeduplicator, RoomDirectory, TicTacToeCoordinator, authorizeRoomCommand, authorizedRoomSnapshot, authorizeTicTacToeCommand, authorizedTicTacToeState, canTransitionRoom, initializeTicTacToe, publishTicTacToeRules, projectRoomID, roomRequest, ticTacToeDraw, ticTacToeWinner, validateMembership, validateTicTacToeCommandPayload, validateTicTacToeMove } from "../../plugins/rooms/index.js";
import { createRouteHandler } from "../../plugins/sdk-typescript/src/index.js";

const request = roomRequest("dev", "rooms", "0.1.0", "rooms.create", "room-1", { maxPlayers: 2 });
if (request.params.pluginId !== "rooms" || request.params.operation !== "rooms.create") throw new Error("rooms plugin contract failed");
if (typeof createRouteHandler([{ operation: "rooms.create", function: "create", capability: "rooms.create" }])(request) !== "object") throw new Error("rooms handler contract failed");

const tickets = new ConnectionTickets();
const ticket = tickets.issue("dev", "room-new", "user-1", 1000, 60);
if (!tickets.consume(ticket.token, "dev", "room-new", "user-1", 1001)) throw new Error("connection ticket was not consumed");
if (tickets.consume(ticket.token, "dev", "room-new", "user-1", 1002)) throw new Error("connection ticket was reusable");
const wrongProject = tickets.issue("dev", "room-new", "user-2", 1000, 60);
if (tickets.consume(wrongProject.token, "prod", "room-new", "user-2", 1001)) throw new Error("cross-project ticket accepted");
if (projectRoomID("dev", "room-1") !== "dev:room-1") throw new Error("project room id was not scoped");
try { projectRoomID("dev", "prod:room"); throw new Error("invalid room id accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "invalid project room id") throw error; }
validateMembership("dev", "room-1", [{ playerId: "user-1", slot: 0 }, { playerId: "user-2", slot: 1 }], 2);
try { validateMembership("dev", "room-1", [{ playerId: "user-1", slot: 0 }, { playerId: "user-2", slot: 0 }], 2); throw new Error("duplicate player slot accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "invalid player slots") throw error; }
if (!canTransitionRoom("waiting", "matching") || !canTransitionRoom("matching", "active") || canTransitionRoom("closed", "active")) throw new Error("room lifecycle transition invalid");
const savedMembership = new Map<string, number>();
const rooms = new RoomDirectory({ save: (roomId, slots) => savedMembership.set(roomId, slots.length) });
if (rooms.create("dev", "room-new", 4).state !== "waiting") throw new Error("room did not start waiting");
try { rooms.create("dev", "room-new", 4); throw new Error("duplicate room accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "room already exists") throw error; }
try { rooms.get("prod", "room-new"); throw new Error("cross-project room lookup accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "room not found") throw error; }
rooms.join("dev", "room-new", "user-1", 0);
if (rooms.reconnect("dev", "room-new", "user-1").slots.length !== 1) throw new Error("existing player did not reconnect");
try { rooms.reconnect("dev", "room-new", "user-2"); throw new Error("unknown player reconnected"); } catch (error) { if (!(error instanceof Error) || error.message !== "player not in room") throw error; }
if (authorizedRoomSnapshot(rooms.get("dev", "room-new"), "dev", "user-1").projectId !== "dev") throw new Error("authorized snapshot missing");
if (!authorizeRoomCommand(rooms.get("dev", "room-new"), "dev", "user-1", "room.action")) throw new Error("authorized command rejected");
try { authorizeRoomCommand(rooms.get("dev", "room-new"), "dev", "user-2", "room.action"); throw new Error("unauthorized command accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "room membership required") throw error; }
try { authorizedRoomSnapshot(rooms.get("dev", "room-new"), "dev", "user-2"); throw new Error("unauthorized snapshot accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "room membership required") throw error; }
if (savedMembership.get("dev:room-new") !== 1) throw new Error("room membership was not persisted");
if (!rooms.isMember("dev", "room-new", "user-1") || rooms.isMember("dev", "room-new", "user-2")) throw new Error("room membership check failed");
try { rooms.join("dev", "room-new", "user-1", 1); throw new Error("duplicate membership accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "invalid membership") throw error; }
rooms.leave("dev", "room-new", "user-1");
if (rooms.events("dev", "room-new").map((event) => event.type).join(",") !== "player.joined,player.left") throw new Error("room events failed");
if (rooms.eventsSince("dev", "room-new", 1).length !== 1 || rooms.eventsSince("dev", "room-new", 1)[0].type !== "player.left") throw new Error("missed events were not replayed");
try { rooms.join("prod", "room-new", "user-2", 1); throw new Error("cross-project join accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "room not found") throw error; }
const game = initializeTicTacToe("dev", "room-1", "user-1", "user-2");
if (game.board.length !== 9 || game.turn !== "user-1" || game.projectId !== "dev") throw new Error("tic-tac-toe initialization failed");
const moved = validateTicTacToeMove(game, "user-1", 0);
if (moved.board[0] !== "X" || moved.turn !== "user-2") throw new Error("legal tic-tac-toe move failed");
try { validateTicTacToeMove(moved, "user-1", 1); throw new Error("out-of-turn move accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "out-of-turn move") throw error; }
try { validateTicTacToeMove(moved, "user-2", 0); throw new Error("occupied tic-tac-toe cell accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "cell occupied") throw error; }
let winningGame = initializeTicTacToe("dev", "room-win", "user-1", "user-2");
winningGame = validateTicTacToeMove(winningGame, "user-1", 0); winningGame = validateTicTacToeMove(winningGame, "user-2", 3);
winningGame = validateTicTacToeMove(winningGame, "user-1", 1); winningGame = validateTicTacToeMove(winningGame, "user-2", 4);
winningGame = validateTicTacToeMove(winningGame, "user-1", 2);
if (ticTacToeWinner(winningGame) !== "user-1") throw new Error("tic-tac-toe win was not detected");
let drawnGame = initializeTicTacToe("dev", "room-draw", "user-1", "user-2");
for (const cell of [0, 1, 2, 4, 3, 5, 7, 6, 8]) drawnGame = validateTicTacToeMove(drawnGame, drawnGame.turn, cell);
if (!ticTacToeDraw(drawnGame) || ticTacToeWinner(drawnGame)) throw new Error("tic-tac-toe draw was not detected");
const rulesArtifact = publishTicTacToeRules((source) => `compiled:${source.length}`);
if (!rulesArtifact.startsWith("compiled:")) throw new Error("tic-tac-toe rules were not compiled");
try { publishTicTacToeRules(() => ""); throw new Error("empty rules artifact accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "rules compilation failed") throw error; }
if (authorizedTicTacToeState(game, "dev", "user-1").roomId !== "room-1" || !authorizeTicTacToeCommand(game, "dev", "user-1", "move")) throw new Error("authorized tic-tac-toe access failed");
try { authorizedTicTacToeState(game, "prod", "user-1"); throw new Error("cross-project game state exposed"); } catch (error) { if (!(error instanceof Error) || error.message !== "game authorization required") throw error; }
const commands = new CommandDeduplicator();
if (!commands.accept("cmd-1") || commands.accept("cmd-1")) throw new Error("duplicate room command accepted");
if (validateTicTacToeCommandPayload({ cell: 4 }).cell !== 4) throw new Error("valid command payload rejected");
try { validateTicTacToeCommandPayload({ cell: 4, extra: true }); throw new Error("mismatched command payload accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "command payload mismatch") throw error; }
const coordinator = new TicTacToeCoordinator(game);
const revision = coordinator.revision();
coordinator.move("user-1", 1, revision);
try { coordinator.move("user-2", 2, revision); throw new Error("concurrent stale move accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "stale game revision") throw error; }
