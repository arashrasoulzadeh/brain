import { handle, inviteRequest, manifest } from "../../plugins/invites/src/index.js";

const request = inviteRequest("dev", "invites.create", "room-1", { target: "user-2" });
if (manifest.id !== "invites" || request.params.operation !== "invites.create") throw new Error("invites plugin contract failed");
if (typeof handle(request) !== "object") throw new Error("invites handler contract failed");
