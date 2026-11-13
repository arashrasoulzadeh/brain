import { AuthService } from "../../plugins/auth/index.js";
import { ArenaStateMachine, arenaStartCommand } from "../../plugins/arena/index.js";

const auth = new AuthService();
const first = auth.guest("dev");
const second = auth.guest("dev");
const room = new ArenaStateMachine("dev", "guest-coop");
arenaStartCommand(room, "dev", "guest-coop", [first.userId, second.userId]);
if (room.getState() !== "active" || room.players.length !== 2) throw new Error("guest co-op rerun failed");
try { auth.session("prod", first.token); throw new Error("guest session crossed project"); } catch (error) { if (!(error instanceof Error) || error.message !== "invalid session") throw error; }
