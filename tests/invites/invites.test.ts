import { InviteAccepter, InviteCodeAccepter, InviteCreator, inviteRequest } from "../../plugins/invites/index.js";
import { createRouteHandler } from "../../plugins/sdk-typescript/src/index.js";

const request = inviteRequest("dev", "invites", "0.1.0", "invites.create", "room-1", { target: "user-2" });
if (request.params.pluginId !== "invites" || request.params.operation !== "invites.create") throw new Error("invites plugin contract failed");
if (typeof createRouteHandler([{ operation: "invites.create", function: "create", capability: "invites.create" }])(request) !== "object") throw new Error("invites handler contract failed");
const inviteMap = new Map(); let inviteSequence = 0; const creator = new InviteCreator(inviteMap, () => ++inviteSequence); const accepter = new InviteAccepter(inviteMap); const codeAccepter = new InviteCodeAccepter(inviteMap);
try { creator.create("dev:other", "room-1", "user-2", 60, 1000); throw new Error("malformed invite project accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "invalid invite") throw error; }
const invite = creator.create("dev", "room-1", "user-2", 60, 1000);
if (!invite.id.startsWith("dev:") || !invite.code.startsWith("INVITE-") || invite.expiresAt !== 61000 || invite.used) throw new Error("invite creation failed");
if (!accepter.accept(invite.id, "dev", "user-2", 2000).used) throw new Error("invite was not consumed");
try { accepter.accept(invite.id, "dev", "user-2", 2000); throw new Error("invite replay accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "invite unavailable") throw error; }
const expired = creator.create("dev", "room-2", "user-3", 1, 1000);
try { accepter.accept(expired.id, "dev", "user-3", 2000); throw new Error("expired invite accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "invite unavailable") throw error; }
const isolated = creator.create("dev", "room-3", "user-4", 60, 1000);
try { accepter.accept(isolated.id, "prod", "user-4", 2000); throw new Error("cross-project invite accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "invite unavailable") throw error; }
const coded = creator.create("dev", "room-4", "user-5", 60, 1000);
if (!codeAccepter.acceptCode(coded.code, "dev", "user-5", 2000).used) throw new Error("invitation code failed");
