import { MatchmakingConfigs, MatchmakingQueue, matchmakingRequest } from "../../plugins/matchmaking/index.js";
import { createRouteHandler } from "../../plugins/sdk-typescript/src/index.js";

const request = matchmakingRequest("dev", "matchmaking", "0.1.0", "matchmaking.queue", "casual", { playerId: "user-1" });
if (request.params.pluginId !== "matchmaking" || request.params.operation !== "matchmaking.queue") throw new Error("matchmaking plugin contract failed");
if (typeof createRouteHandler([{ operation: "matchmaking.queue", function: "queue", capability: "matchmaking.queue" }])(request) !== "object") throw new Error("matchmaking handler contract failed");
const configs = new MatchmakingConfigs();
configs.set({ projectId: "dev", scope: "ranked", matchSize: 2, ticketTtlSeconds: 60, enabled: true });
if (configs.get("dev", "ranked").matchSize !== 2) throw new Error("matchmaking config was not stored");
try { configs.get("prod", "ranked"); throw new Error("cross-project matchmaking config accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "matchmaking config not found") throw error; }
const queue = new MatchmakingQueue();
const first = queue.enqueue("dev", "ranked", "user-1", 1000);
queue.enqueue("dev", "ranked", "user-2", 2000);
if (queue.peek("dev", "ranked")[0].id !== first.id || queue.peek("prod", "ranked").length !== 0) throw new Error("project queue isolation or FIFO failed");
