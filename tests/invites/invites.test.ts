import { InviteDirectory, inviteRequest } from "../../plugins/invites/index.js";
import { createRouteHandler } from "../../plugins/sdk-typescript/src/index.js";

const request = inviteRequest("dev", "invites", "0.1.0", "invites.create", "room-1", { target: "user-2" });
if (request.params.pluginId !== "invites" || request.params.operation !== "invites.create") throw new Error("invites plugin contract failed");
if (typeof createRouteHandler([{ operation: "invites.create", function: "create", capability: "invites.create" }])(request) !== "object") throw new Error("invites handler contract failed");
const invites = new InviteDirectory();
const invite = invites.create("dev", "room-1", "user-2", 60, 1000);
if (!invite.id.startsWith("dev:") || invite.expiresAt !== 61000 || invite.used) throw new Error("invite creation failed");
