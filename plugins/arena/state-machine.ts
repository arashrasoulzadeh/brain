import { createArenaRuntime } from "./arena-state.js";
import { ArenaActionRecorder } from "./arena-action-recorder.js";
import { ArenaFinisher } from "./arena-finisher.js";
import { ArenaMeasurer } from "./arena-measurer.js";
import { ArenaStarter } from "./arena-starter.js";
import { ArenaStateReader } from "./arena-state-reader.js";
export function createArena(projectId: string, roomId: string) { const runtime = createArenaRuntime(projectId, roomId); const reader = new ArenaStateReader(runtime); const starter = new ArenaStarter(runtime); const actions = new ArenaActionRecorder(runtime); const measurer = new ArenaMeasurer(runtime); const finisher = new ArenaFinisher(runtime); return { projectId, roomId, getState: reader.read.bind(reader), start: starter.start.bind(starter), action: actions.record.bind(actions), measure: measurer.measure.bind(measurer), result: finisher.finish.bind(finisher), get players() { return runtime.players; } }; }
