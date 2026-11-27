import { projectRoomID, nextTurn, createRevisionRuntime, commitRevision, type RevisionRuntime } from "../sdk-typescript/src/index.js";
import { type TicTacToeState } from "./index.js";

export function initializeTicTacToe(
  projectId: string,
  roomId: string,
  firstPlayer: string,
  secondPlayer: string,
): TicTacToeState {
  projectRoomID(projectId, roomId);
  if (!firstPlayer.trim() || !secondPlayer.trim() || firstPlayer === secondPlayer)
    throw new Error("invalid tic-tac-toe players");
  return {
    projectId,
    roomId,
    players: [firstPlayer, secondPlayer],
    board: Array(9).fill(null),
    turn: firstPlayer,
    status: "active",
  };
}

export function validateTicTacToeMove(
  state: TicTacToeState,
  playerId: string,
  cell: number,
): TicTacToeState {
  if (
    state.status !== "active" ||
    !state.players.includes(playerId) ||
    !Number.isInteger(cell) ||
    cell < 0 ||
    cell >= 9
  )
    throw new Error("invalid tic-tac-toe move");
  if (state.turn !== playerId) throw new Error("out-of-turn move");
  if (state.board[cell] !== null) throw new Error("cell occupied");
  const mark = state.players[0] === playerId ? "X" : "O";
  const board = [...state.board] as ("X" | "O" | null)[];
  board[cell] = mark;
  return { ...state, board, turn: nextTurn(state.players, playerId) };
}

export function ticTacToeWinner(state: TicTacToeState): string | undefined {
  const lines = [
    [0, 1, 2],
    [3, 4, 5],
    [6, 7, 8],
    [0, 3, 6],
    [1, 4, 7],
    [2, 5, 8],
    [0, 4, 8],
    [2, 4, 6],
  ];
  for (const [a, b, c] of lines)
    if (state.board[a] && state.board[a] === state.board[b] && state.board[a] === state.board[c])
      return state.players[state.board[a] === "X" ? 0 : 1];
  return undefined;
}

export function ticTacToeDraw(state: TicTacToeState): boolean {
  return !ticTacToeWinner(state) && state.board.every((cell) => cell !== null);
}

export function validateTicTacToeRuleResult(value: unknown): {
  board: readonly ("X" | "O" | null)[];
  turn: string;
  status: "active" | "paused" | "finished";
} {
  if (
    !value ||
    typeof value !== "object" ||
    Object.keys(value).some((key) => !["board", "turn", "status"].includes(key))
  )
    throw new Error("malformed rule result");
  const result = value as { board?: unknown; turn?: unknown; status?: unknown };
  if (
    !Array.isArray(result.board) ||
    result.board.length !== 9 ||
    result.board.some((cell) => cell !== null && cell !== "X" && cell !== "O") ||
    typeof result.turn !== "string" ||
    !["active", "paused", "finished"].includes(String(result.status))
  )
    throw new Error("malformed rule result");
  return {
    board: [...result.board] as ("X" | "O" | null)[],
    turn: result.turn,
    status: result.status as "active" | "paused" | "finished",
  };
}

export const TIC_TAC_TOE_RULES_SOURCE = "package main\n// Brain Tic-tac-toe rules\n";
export function publishTicTacToeRules(compile: (source: string) => string): string {
  const artifact = compile(TIC_TAC_TOE_RULES_SOURCE);
  if (!artifact) throw new Error("rules compilation failed");
  return artifact;
}

export type TicTacToeRuntime = RevisionRuntime<TicTacToeState>;

export function copyTicTacToeState(state: TicTacToeState): TicTacToeState {
  return { ...state, players: [...state.players] as [string, string], board: [...state.board] };
}

export function createTicTacToeRuntime(state: TicTacToeState): TicTacToeRuntime {
  return createRevisionRuntime(state, copyTicTacToeState);
}

export class TicTacToeMover {
  constructor(private readonly runtime: TicTacToeRuntime) {}
  move(playerId: string, cell: number, expectedRevision: number) {
    return commitRevision(
      this.runtime,
      expectedRevision,
      (state) => validateTicTacToeMove(state, playerId, cell),
      copyTicTacToeState,
    );
  }
}

export class TicTacToeRevisionReader {
  constructor(private readonly runtime: TicTacToeRuntime) {}
  read(): number {
    return this.runtime.revision;
  }
}

export function createTicTacToeCoordinator(state: TicTacToeState) {
  const runtime = createTicTacToeRuntime(state);
  const revision = new TicTacToeRevisionReader(runtime);
  const mover = new TicTacToeMover(runtime);
  return { revision: revision.read.bind(revision), move: mover.move.bind(mover) };
}
