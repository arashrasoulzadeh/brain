import { MatchmakingConfigReader, MatchmakingConfigStatus, MatchmakingConfigWriter, createMatchmakingQueue, matchmakingRequest } from "../../plugins/matchmaking/index.js";
import { createRouteHandler } from "../../plugins/sdk-typescript/src/index.js";

const request = matchmakingRequest("dev", "matchmaking", "0.1.0", "matchmaking.queue", "casual", { playerId: "user-1" });
if (request.params.pluginId !== "matchmaking" || request.params.operation !== "matchmaking.queue") throw new Error("matchmaking plugin contract failed");
if (typeof createRouteHandler([{ operation: "matchmaking.queue", function: "queue", capability: "matchmaking.queue" }])(request) !== "object") throw new Error("matchmaking handler contract failed");
const configMap = new Map(); const configWriter = new MatchmakingConfigWriter(configMap); const configReader = new MatchmakingConfigReader(configMap); const configStatus = new MatchmakingConfigStatus(configMap);
configWriter.set({ projectId: "dev", scope: "ranked", matchSize: 2, ticketTtlSeconds: 60, enabled: true });
if (configReader.get("dev", "ranked").matchSize !== 2) throw new Error("matchmaking config was not stored");
try { configReader.get("prod", "ranked"); throw new Error("cross-project matchmaking config accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "matchmaking config not found") throw error; }
const queue = createMatchmakingQueue();
const first = queue.enqueue("dev", "ranked", "user-1", 1000);
queue.enqueue("dev", "ranked", "user-2", 2000);
if (queue.peek("dev", "ranked")[0].id !== first.id || queue.peek("prod", "ranked").length !== 0 || first.expiresAt !== 61000) throw new Error("project queue isolation or FIFO failed");
try { queue.enqueue("dev", "ranked", "user-1", 3000); throw new Error("duplicate active ticket accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "active ticket already exists") throw error; }
const matched = queue.match("dev", "ranked", 2);
if (matched.length !== 2 || matched[0].playerId !== "user-1" || queue.peek("dev", "ranked").length !== 0) throw new Error("compatible FIFO matching failed");
const disabledMap = new Map(); const disabledWriter = new MatchmakingConfigWriter(disabledMap); disabledWriter.set({ projectId: "dev", scope: "casual", matchSize: 2, ticketTtlSeconds: 60, enabled: false });
try { createMatchmakingQueue({ reader: new MatchmakingConfigReader(disabledMap), status: new MatchmakingConfigStatus(disabledMap) }).enqueue("dev", "casual", "user-3"); throw new Error("disabled scope accepted ticket"); } catch (error) { if (!(error instanceof Error) || error.message !== "matchmaking scope disabled") throw error; }
const expiringQueue = createMatchmakingQueue();
expiringQueue.enqueue("dev", "ranked", "user-4", 1000);
expiringQueue.enqueue("dev", "ranked", "user-5", 2000);
if (expiringQueue.expire(62000).length !== 2 || expiringQueue.peek("dev", "ranked").length !== 0) throw new Error("ticket expiry failed");
const retryQueue = createMatchmakingQueue();
const failed = retryQueue.enqueue("dev", "ranked", "user-6", 1000);
retryQueue.match("dev", "ranked", 1);
retryQueue.requeue(failed, 2000);
if (retryQueue.peek("dev", "ranked")[0].id !== failed.id) throw new Error("failed assignment was not requeued");
const assignmentQueue = createMatchmakingQueue();
assignmentQueue.enqueue("dev", "ranked", "user-7", 1000);
assignmentQueue.enqueue("dev", "ranked", "user-8", 1100);
const decision = assignmentQueue.assign("dev", "ranked", 2, 2000);
if (!decision || decision.ticketIds.length !== 2 || !decision.id.startsWith("dev:ranked:assignment-")) throw new Error("assignment decision was not recorded");
