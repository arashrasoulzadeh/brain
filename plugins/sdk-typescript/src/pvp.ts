import { validateProjectId } from "./validation.js";

export type TurnMatchState = {
  projectId: string;
  roomId: string;
  players: readonly string[];
  turn: string;
  status: "active" | "paused";
  turnDeadlineAt?: number;
};

export type RevisionRuntime<S> = { current: S; revision: number };
export type CommittedCommand = { revision: number; playerId: string };

export class CommandDeduplicator {
  private readonly commands = new Set<string>();
  accept(commandId: string): boolean {
    if (!commandId.trim()) throw new Error("command id required");
    if (this.commands.has(commandId)) return false;
    this.commands.add(commandId);
    return true;
  }
}

export function projectRoomID(projectId: string, roomId: string): string {
  validateProjectId(projectId);
  validateProjectId(roomId, "invalid project room id");
  return `${projectId}:${roomId}`;
}

export function assertMatchMember(
  state: Pick<TurnMatchState, "projectId" | "players">,
  projectId: string,
  playerId: string,
): void {
  validateProjectId(projectId, "game authorization required");
  if (state.projectId !== projectId || !Array.isArray(state.players) || !state.players.includes(playerId))
    throw new Error("game authorization required");
}

export function nextTurn(players: readonly string[], current: string): string {
  const index = players.indexOf(current);
  if (index < 0 || players.length < 2) throw new Error("invalid game turn");
  return players[(index + 1) % players.length];
}

export function createRevisionRuntime<S>(state: S, copy: (state: S) => S): RevisionRuntime<S> {
  return { current: copy(state), revision: 0 };
}

export function commitRevision<S>(
  runtime: RevisionRuntime<S>,
  expectedRevision: number,
  apply: (state: S) => S,
  copy: (state: S) => S,
): S {
  if (expectedRevision !== runtime.revision) throw new Error("stale game revision");
  const next = apply(copy(runtime.current));
  const snapshot = copy(next);
  runtime.current = copy(next);
  runtime.revision += 1;
  return snapshot;
}

export function replayHistory<S, C extends CommittedCommand>(
  initial: S,
  commands: readonly C[],
  apply: (state: S, command: C) => S,
): S {
  return commands.reduce((state, command, revision) => {
    if (command.revision !== revision || !command.playerId.trim())
      throw new Error("invalid committed game history");
    try {
      return apply(state, command);
    } catch {
      throw new Error("invalid committed game history");
    }
  }, initial);
}

export function expireTurn<S extends TurnMatchState>(
  state: S,
  now: number,
  nextDeadlineAt: number,
  errors: { invalid: string; unexpired: string },
): S {
  if (state.status !== "active" || !Number.isFinite(now) ||
      !Number.isFinite(nextDeadlineAt) || nextDeadlineAt <= now)
    throw new Error(errors.invalid);
  if (state.turnDeadlineAt === undefined || !Number.isFinite(state.turnDeadlineAt) || now < state.turnDeadlineAt)
    throw new Error(errors.unexpired);
  return { ...state, turn: nextTurn(state.players, state.turn), turnDeadlineAt: nextDeadlineAt };
}

export function pauseTurn<S extends TurnMatchState>(
  state: S,
  copy: (state: S) => S,
  error: string,
): S {
  if (state.status !== "active") throw new Error(error);
  return { ...copy(state), status: "paused", turnDeadlineAt: undefined };
}

export function resumeTurn<S extends TurnMatchState>(
  state: S,
  requestedVersion: string,
  pinnedVersion: string,
  copy: (state: S) => S,
  error: string,
): S {
  if (state.status !== "paused") throw new Error(error);
  if (!requestedVersion || requestedVersion !== pinnedVersion)
    throw new Error("pinned room version required");
  return { ...copy(state), status: "active" };
}

export function privatePlayerView<S extends TurnMatchState, T>(
  state: S,
  playerId: string,
  privateState: Readonly<Record<string, T>>,
  copy: (state: S) => S,
): S & { privateState?: T } {
  return {
    ...copy(state),
    privateState: state.players.includes(playerId) && Object.hasOwn(privateState, playerId)
      ? privateState[playerId]
      : undefined,
  };
}

export function parseMatchState(serialized: string): Record<string, unknown> {
  let parsed: unknown;
  try {
    parsed = JSON.parse(serialized);
  } catch {
    throw new Error("invalid persisted game");
  }
  if (!parsed || typeof parsed !== "object" || Array.isArray(parsed))
    throw new Error("invalid persisted game");
  return parsed as Record<string, unknown>;
}
