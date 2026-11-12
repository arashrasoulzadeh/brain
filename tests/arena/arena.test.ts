import { ArenaStateMachine, arenaRequest, validateMatchResult } from "../../plugins/arena/index.js";
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
try { validateMatchResult({ resultId: "prod:result-1", projectId: "dev", roomId: "room-1", winnerIds: ["user-1"], completedAt: new Date().toISOString() }); throw new Error("cross-project result accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "result project mismatch") throw error; }
