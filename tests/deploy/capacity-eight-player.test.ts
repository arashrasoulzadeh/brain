import { createArena } from "../../plugins/arena/index.js";

const players = Array.from({ length: 8 }, (_, index) => `capacity-player-${index + 1}`);
const arena = createArena("capacity", "reference-eight-player");
arena.start(players);
if (arena.players.length !== 8 || arena.getState() !== "active")
  throw new Error("reference eight-player room did not start");
const measurement = arena.measure(10);
if (measurement.iterations !== 10 || measurement.elapsedMs < 0 || measurement.memoryBytes <= 0)
  throw new Error("reference eight-player measurement was invalid");
