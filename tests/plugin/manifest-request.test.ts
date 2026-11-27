import { deepEqual, equal } from "node:assert/strict";
import { createManifestRequest, type ManifestOperation } from "../../plugins/sdk-typescript/src/index.js";
import { createMatchmakingQueue } from "../../plugins/matchmaking/index.js";

const operation: ManifestOperation<"rooms"> = "rooms.join";
const make = createManifestRequest("rooms", "roomId");
const value = make("alpha", "1.0.0", operation, "trusted", { roomId: "forged", name: "test" });
equal(value.params.pluginId, "rooms");
equal(value.params.projectId, "alpha");
equal(value.params.operation, "rooms.join");
deepEqual(value.params.payload, { roomId: "trusted", name: "test" });

const queue = createMatchmakingQueue();
queue.enqueue("alpha", "scope", "player", 0);
const first = queue.peek("alpha", "scope")[0];
first.playerId = "changed";
equal(queue.peek("alpha", "scope")[0].playerId, "player");
deepEqual(queue.match("alpha", "scope", 0.5), []);
equal(queue.assign("alpha", "scope", 0), null);
equal(queue.match("alpha", "scope", 1).length, 1);
