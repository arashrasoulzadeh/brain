import { InviteDirectory, inviteRequest } from "../../plugins/invites/index.js";
import { createRouteHandler } from "../../plugins/sdk-typescript/src/index.js";

const request = inviteRequest("dev", "invites", "0.1.0", "invites.create", "room-1", { target: "user-2" });
if (request.params.pluginId !== "invites" || request.params.operation !== "invites.create") throw new Error("invites plugin contract failed");
if (typeof createRouteHandler([{ operation: "invites.create", function: "create", capability: "invites.create" }])(request) !== "object") throw new Error("invites handler contract failed");
const invites = new InviteDirectory();
const invite = invites.create("dev", "room-1", "user-2", 60, 1000);
if (!invite.id.startsWith("dev:") || !invite.code.startsWith("INVITE-") || invite.expiresAt !== 61000 || invite.used) throw new Error("invite creation failed");
if (!invites.accept(invite.id, "dev", "user-2", 2000).used) throw new Error("invite was not consumed");
try { invites.accept(invite.id, "dev", "user-2", 2000); throw new Error("invite replay accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "invite unavailable") throw error; }
const expired = invites.create("dev", "room-2", "user-3", 1, 1000);
try { invites.accept(expired.id, "dev", "user-3", 2000); throw new Error("expired invite accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "invite unavailable") throw error; }
const isolated = invites.create("dev", "room-3", "user-4", 60, 1000);
try { invites.accept(isolated.id, "prod", "user-4", 2000); throw new Error("cross-project invite accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "invite unavailable") throw error; }
const coded = invites.create("dev", "room-4", "user-5", 60, 1000);
if (!invites.acceptCode(coded.code, "dev", "user-5", 2000).used) throw new Error("invitation code failed");
