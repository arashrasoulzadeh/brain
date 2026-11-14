import { createArena } from "../../plugins/arena/index.js";
import { wasmBuildSettings } from "../../plugins/functions/index.js";

const compilerStarted = performance.now();
const settings = wasmBuildSettings();
const compilerTimeMs = performance.now() - compilerStarted;
if (settings.goos !== "wasip1" || compilerTimeMs < 0) throw new Error("compiler metric missing");
const arena = createArena("dev", "metrics-room");
arena.start(["player-a", "player-b"]);
const turnStarted = performance.now();
arena.action("player-a", 1, { cell: 0 });
const turnLatencyMs = performance.now() - turnStarted;
const measurement = arena.measure(10);
if (turnLatencyMs < 0 || measurement.elapsedMs < 0 || measurement.memoryBytes < 1)
  throw new Error("turn or memory metric missing");
