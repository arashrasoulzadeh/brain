import { captureRequest, validateEvent } from "../../plugins/analytics/index.js";

validateEvent({ projectId: "dev", name: "room.joined", fields: { roomId: "room-1", players: 2 } });
const request = captureRequest("dev", "analytics", "0.1.0", { name: "room.joined", fields: { roomId: "room-1" } });
if (request.params.operation !== "analytics.capture") throw new Error("analytics route contract failed");
try { validateEvent({ projectId: "dev", name: "room joined", fields: {} }); throw new Error("invalid event name accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "invalid event name") throw error; }
try { validateEvent({ projectId: "dev", name: "room.joined", fields: { "bad-key": "x" } }); throw new Error("invalid event field accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "invalid event fields") throw error; }
