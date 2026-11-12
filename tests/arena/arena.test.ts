import { ArenaStateMachine, MatchResultStore, arenaRequest, arenaStartCommand, exportEyeEvent, validateMatchResult } from "../../plugins/arena/index.js";
import { createRouteHandler } from "../../plugins/sdk-typescript/src/index.js";

const request = arenaRequest("dev", "arena", "0.1.0", "arena.start", "room-1", { players: ["user-1", "user-2"] });
if (request.params.pluginId !== "arena" || request.params.operation !== "arena.start") throw new Error("arena plugin contract failed");
if (typeof createRouteHandler([{ operation: "arena.start", function: "start", capability: "arena.start" }])(request) !== "object") throw new Error("arena handler contract failed");
validateMatchResult({ resultId: "dev:result-1", projectId: "dev", roomId: "room-1", winnerIds: ["user-1"], completedAt: new Date().toISOString() });
const arena = new ArenaStateMachine("dev", "room-1");
if (arena.getState() !== "waiting") throw new Error("arena did not start waiting");
arena.start(["user-1", "user-2"]);
if (arena.getState() !== "active" || arena.players.length !== 2) throw new Error("arena did not become active");
try { arena.start(["user-1"]); throw new Error("active arena restarted"); } catch (error) { if (!(error instanceof Error) || error.message !== "arena already started") throw error; }
const commandArena = new ArenaStateMachine("dev", "room-2");
if (arenaStartCommand(commandArena, "dev", "room-2", ["user-1", "user-2"]).getState() !== "active") throw new Error("start command failed");
commandArena.action("user-1", 1, { move: "left" });
try { commandArena.action("user-1", 1, { move: "right" }); throw new Error("duplicate action accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "action sequence conflict") throw error; }
try { commandArena.action("user-9", 1, {}); throw new Error("unknown player action accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "player not in arena") throw error; }
const result = commandArena.result(["user-1"], "2026-11-13T00:00:00.000Z");
if (commandArena.getState() !== "finished" || result.winnerIds[0] !== "user-1") throw new Error("result command failed");
const resultStore = new MatchResultStore();
resultStore.save(result);
if (resultStore.get(result.resultId)?.roomId !== "room-2") throw new Error("completed result was not persisted");
if (exportEyeEvent(result).name !== "arena.result") throw new Error("Eye event was not exported");
try { resultStore.save({ ...result, roomId: "room-other" }); throw new Error("result overwrite accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "result already exists") throw error; }
try { arenaStartCommand(commandArena, "prod", "room-2", ["user-1", "user-2"]); throw new Error("cross-project start accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "arena project mismatch") throw error; }
try { validateMatchResult({ resultId: "prod:result-1", projectId: "dev", roomId: "room-1", winnerIds: ["user-1"], completedAt: new Date().toISOString() }); throw new Error("cross-project result accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "result project mismatch") throw error; }
