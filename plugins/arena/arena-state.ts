import type { ArenaState } from "./types.js";
export type ArenaRuntime = {
  projectId: string;
  roomId: string;
  state: ArenaState;
  players: string[];
  sequences: Map<string, number>;
};
export function createArenaRuntime(projectId: string, roomId: string): ArenaRuntime {
  return { projectId, roomId, state: "waiting", players: [], sequences: new Map() };
}
