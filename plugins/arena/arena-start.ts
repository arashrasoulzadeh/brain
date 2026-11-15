import type { createArena } from "./state-machine.js";
import { validateProjectId } from "../sdk-typescript/src/index.js";

export function arenaStartCommand(
  arena: ReturnType<typeof createArena>,
  projectId: string,
  roomId: string,
  players: readonly string[],
): ReturnType<typeof createArena> {
  try {
    validateProjectId(projectId, "arena project mismatch");
  } catch {
    throw new Error("arena project mismatch");
  }
  if (arena.projectId !== projectId || arena.roomId !== roomId)
    throw new Error("arena project mismatch");
  arena.start(players);
  return arena;
}
