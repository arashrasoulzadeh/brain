import type { createArena } from "./state-machine.js";

export function arenaStartCommand(
  arena: ReturnType<typeof createArena>,
  projectId: string,
  roomId: string,
  players: readonly string[],
): ReturnType<typeof createArena> {
  if (arena.projectId !== projectId || arena.roomId !== roomId)
    throw new Error("arena project mismatch");
  arena.start(players);
  return arena;
}
