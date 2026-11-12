import { createHandler, inviteRequest } from "../../plugins/invites/index.js";

const request = inviteRequest("dev", "invites", "0.1.0", "invites.create", "room-1", { target: "user-2" });
if (request.params.pluginId !== "invites" || request.params.operation !== "invites.create") throw new Error("invites plugin contract failed");
if (typeof createHandler([{ operation: "invites.create", function: "create", capability: "invites.create" }])(request) !== "object") throw new Error("invites handler contract failed");
