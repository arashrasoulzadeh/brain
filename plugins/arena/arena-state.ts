import type { ArenaState } from "./types.js";
import { validateProjectId } from "../sdk-typescript/src/index.js";
export type ArenaRuntime = {
  projectId: string;
  roomId: string;
  state: ArenaState;
  players: string[];
  sequences: Map<string, number>;
};
export function createArenaRuntime(projectId: string, roomId: string): ArenaRuntime {
  validateProjectId(projectId, "invalid arena project");
  validateProjectId(roomId, "invalid arena room");
  return { projectId, roomId, state: "waiting", players: [], sequences: new Map() };
}
