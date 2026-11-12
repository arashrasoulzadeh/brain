import { ArenaStateMachine, arenaStartCommand } from "../../plugins/arena/index.js";

const browserOne = new ArenaStateMachine("dev", "gate-room");
const browserTwo = new ArenaStateMachine("dev", "gate-room");
arenaStartCommand(browserOne, "dev", "gate-room", ["user-1", "user-2"]);
arenaStartCommand(browserTwo, "dev", "gate-room", ["user-1", "user-2"]);
if (browserOne.getState() !== "active" || browserTwo.getState() !== "active") throw new Error("two-browser match did not start");
