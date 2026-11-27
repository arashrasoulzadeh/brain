import type { TicTacToeState, CommittedTicTacToeMove } from "./index.js";
import { copyTicTacToeState, validateTicTacToeMove, validateTicTacToeRuleResult } from "./tic-tac-toe.js";
import { assertMatchMember, expireTurn, pauseTurn, resumeTurn, replayHistory, parseMatchState, privatePlayerView, validateProjectId } from "../sdk-typescript/src/index.js";

export function expireTicTacToeTurn(
  state: TicTacToeState,
  now: number,
  nextDeadlineAt: number,
): TicTacToeState {
  return expireTurn(state, now, nextDeadlineAt, {
    invalid: "invalid tic-tac-toe timeout",
    unexpired: "tic-tac-toe turn not expired",
  });
}

export function pauseTicTacToeTurn(state: TicTacToeState): TicTacToeState {
  return pauseTurn(state, copyTicTacToeState, "tic-tac-toe turn is not active");
}

export function resumePausedTicTacToeTurn(
  state: TicTacToeState,
  requestedVersion: string,
  pinnedVersion: string,
): TicTacToeState {
  return resumeTurn(state, requestedVersion, pinnedVersion, copyTicTacToeState, "tic-tac-toe turn is not paused");
}

export function reconstructTicTacToeHistory(
  initialState: TicTacToeState,
  committedMoves: readonly CommittedTicTacToeMove[],
): TicTacToeState {
  return replayHistory(initialState, committedMoves,
    (state, outcome) => validateTicTacToeMove(state, outcome.playerId, outcome.cell));
}

export function reconnectTicTacToe(
  serialized: string,
  projectId: string,
  playerId: string,
): TicTacToeState {
  validateProjectId(projectId, "game authorization required");
  const state = parseMatchState(serialized) as Partial<TicTacToeState>;
  assertMatchMember(state as TicTacToeState, projectId, playerId);
  const result = validateTicTacToeRuleResult({ board: state.board, turn: state.turn, status: state.status });
  return {
    projectId: state.projectId!,
    roomId: state.roomId!,
    players: [...state.players!] as [string, string],
    board: [...result.board],
    turn: result.turn,
    status: state.status!,
    ...(state.turnDeadlineAt !== undefined ? { turnDeadlineAt: state.turnDeadlineAt } : {}),
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
  return privatePlayerView(state, playerId, privateState, copyTicTacToeState);
}

export function authorizedTicTacToeState(
  state: TicTacToeState,
  projectId: string,
  playerId: string,
): TicTacToeState {
  assertMatchMember(state, projectId, playerId);
  return copyTicTacToeState(state);
}

export function authorizeTicTacToeCommand(
  state: TicTacToeState,
  projectId: string,
  playerId: string,
  operation: string,
): boolean {
  authorizedTicTacToeState(state, projectId, playerId);
  if (operation !== "move") throw new Error("game command not authorized");
  return true;
}

export function validateTicTacToeCommandPayload(payload: unknown): { cell: number } {
  if (
    !payload ||
    typeof payload !== "object" ||
    Object.keys(payload).length !== 1 ||
    typeof (payload as { cell?: unknown }).cell !== "number" ||
    !Number.isInteger((payload as { cell: number }).cell) ||
    (payload as { cell: number }).cell < 0 ||
    (payload as { cell: number }).cell > 8
  )
    throw new Error("command payload mismatch");
  return { cell: (payload as { cell: number }).cell };
}
