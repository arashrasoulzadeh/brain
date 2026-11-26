import { deepEqual, throws } from "node:assert/strict";
import { createResourceRequest, validateBoundedInteger, requireAllowedValues } from "../../plugins/sdk-typescript/src/index.js";
import { fileRequest } from "../../plugins/files/index.js";
import { roomRequest } from "../../plugins/rooms/index.js";

const build = createResourceRequest<"files.upload">("files", "objectId");
deepEqual(build("p", "files", "1", "files.upload", "x", { data: "abc" }), fileRequest("p", "files", "1", "files.upload", "x", { data: "abc" }));
deepEqual(roomRequest("p", "rooms", "1", "rooms.join", "r"), {
  jsonrpc: "2.0", id: "rooms-rooms.join-p-r", method: "invoke",
  params: { protocol: 1, projectId: "p", pluginId: "rooms", version: "1", operation: "rooms.join", payload: { roomId: "r" } },
});
validateBoundedInteger(0, 0, "limit");
validateBoundedInteger(1, 1, "limit", 1);
for (const value of [-1, 1.5, Infinity, 2]) throws(() => validateBoundedInteger(value, 1, "limit"), /limit/);
throws(() => validateBoundedInteger(0, 0, "limit", 1), /limit/);
requireAllowedValues(["ok"], ["ok", "other"], "denied");
throws(() => requireAllowedValues(["bad"], ["ok"], "denied"), /denied/);
