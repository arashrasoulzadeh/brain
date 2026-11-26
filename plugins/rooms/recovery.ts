import { type TicTacToeState, type CommittedTicTacToeMove } from "./index.js";
import { validateTicTacToeMove, validateTicTacToeRuleResult } from "./gameplay.js";
import { validateProjectId } from "../sdk-typescript/src/index.js";

export function expireTicTacToeTurn(
  state: TicTacToeState,
  now: number,
  nextDeadlineAt: number,
): TicTacToeState {
  if (
    state.status !== "active" ||
    !Number.isFinite(now) ||
    !Number.isFinite(nextDeadlineAt) ||
    nextDeadlineAt <= now
  )
    throw new Error("invalid tic-tac-toe timeout");
  if (state.turnDeadlineAt === undefined || now < state.turnDeadlineAt)
    throw new Error("tic-tac-toe turn not expired");
  return {
    ...state,
    turn: state.players.find((player) => player !== state.turn)!,
    turnDeadlineAt: nextDeadlineAt,
  };
}

export function pauseTicTacToeTurn(state: TicTacToeState): TicTacToeState {
  if (state.status !== "active") throw new Error("tic-tac-toe turn is not active");
  return {
    ...state,
    board: [...state.board],
    players: [...state.players] as [string, string],
    status: "paused",
    turnDeadlineAt: undefined,
  };
}

export function resumePausedTicTacToeTurn(
  state: TicTacToeState,
  requestedVersion: string,
  pinnedVersion: string,
): TicTacToeState {
  if (state.status !== "paused") throw new Error("tic-tac-toe turn is not paused");
  if (!requestedVersion || requestedVersion !== pinnedVersion)
    throw new Error("pinned room version required");
  return {
    ...state,
    board: [...state.board],
    players: [...state.players] as [string, string],
    status: "active",
  };
}

export function reconstructTicTacToeHistory(
  initialState: TicTacToeState,
  committedMoves: readonly CommittedTicTacToeMove[],
): TicTacToeState {
  return committedMoves.reduce((state, outcome, expectedRevision) => {
    if (outcome.revision !== expectedRevision || !outcome.playerId.trim())
      throw new Error("invalid committed game history");
    try {
      return validateTicTacToeMove(state, outcome.playerId, outcome.cell);
    } catch {
      throw new Error("invalid committed game history");
    }
  }, initialState);
}

export function reconnectTicTacToe(
  serialized: string,
  projectId: string,
  playerId: string,
): TicTacToeState {
  validateProjectId(projectId, "game authorization required");
  let parsed: unknown;
  try {
    parsed = JSON.parse(serialized);
  } catch {
    throw new Error("invalid persisted game");
  }
  const state = parsed as Partial<TicTacToeState>;
  if (
    state.projectId !== projectId ||
    !Array.isArray(state.players) ||
    !state.players.includes(playerId)
  )
    throw new Error("game authorization required");
  validateTicTacToeRuleResult({ board: state.board, turn: state.turn, status: state.status });
  return {
    projectId: state.projectId,
    roomId: state.roomId!,
    players: [...state.players] as [string, string],
    board: [...state.board!] as ("X" | "O" | null)[],
    turn: state.turn!,
    status: state.status!,
  };
}

export function serializeTicTacToeState(state: TicTacToeState): string {
  return JSON.stringify(state);
}

export function hiddenTicTacToeView<T>(
  state: TicTacToeState,
  playerId: string,
  privateState: Readonly<Record<string, T>>,
): TicTacToeState & { privateState?: T } {
  const view = {
    ...state,
    players: [...state.players] as [string, string],
    board: [...state.board],
  };
  return {
    ...view,
    privateState: state.players.includes(playerId) ? privateState[playerId] : undefined,
  };
}
