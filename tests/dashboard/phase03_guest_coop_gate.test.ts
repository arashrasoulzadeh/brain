import { createAuthOperations } from "../../plugins/auth/index.js";
import { createArena, arenaStartCommand } from "../../plugins/arena/index.js";

const auth = createAuthOperations();
const first = auth.guest("dev");
const second = auth.guest("dev");
const room = createArena("dev", "guest-coop");
arenaStartCommand(room, "dev", "guest-coop", [first.userId, second.userId]);
if (room.getState() !== "active" || room.players.length !== 2)
  throw new Error("guest co-op rerun failed");
try {
  auth.session("prod", first.token);
  throw new Error("guest session crossed project");
} catch (error) {
  if (!(error instanceof Error) || error.message !== "invalid session") throw error;
}
