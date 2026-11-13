import type { ArenaStateMachine } from "./state-machine.js";

export function arenaStartCommand(arena: ArenaStateMachine, projectId: string, roomId: string, players: readonly string[]): ArenaStateMachine {
  if (arena.projectId !== projectId || arena.roomId !== roomId) throw new Error("arena project mismatch");
  arena.start(players);
  return arena;
}
