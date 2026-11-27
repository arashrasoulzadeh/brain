import {
  createTicTacToeCoordinator,
  authorizeTicTacToeCommand,
  authorizedTicTacToeState,
  expireTicTacToeTurn,
  hiddenTicTacToeView,
  initializeTicTacToe,
  pauseTicTacToeTurn,
  publishTicTacToeRules,
  reconstructTicTacToeHistory,
  reconnectTicTacToe,
  resumePausedTicTacToeTurn,
  serializeTicTacToeState,
  ticTacToeDraw,
  ticTacToeWinner,
  validateTicTacToeCommandPayload,
  validateTicTacToeMove,
  validateTicTacToeRuleResult,
} from "../../plugins/games/index.js";

const game = initializeTicTacToe("dev", "room-1", "user-1", "user-2");
if (game.board.length !== 9 || game.turn !== "user-1" || game.projectId !== "dev")
  throw new Error("tic-tac-toe initialization failed");
const timedGame = { ...game, turnDeadlineAt: 1000 };
if (expireTicTacToeTurn(timedGame, 1000, 2000).turn !== "user-2")
  throw new Error("expired turn did not advance");
try {
  expireTicTacToeTurn(timedGame, 999, 2000);
  throw new Error("unexpired turn advanced");
} catch (error) {
  if (!(error instanceof Error) || error.message !== "tic-tac-toe turn not expired") throw error;
}
const reconstructed = reconstructTicTacToeHistory(game, [
  { revision: 0, playerId: "user-1", cell: 0 },
  { revision: 1, playerId: "user-2", cell: 4 },
]);
if (reconstructed.board[0] !== "X" || reconstructed.board[4] !== "O")
  throw new Error("committed game history was not reconstructed");
try {
  reconstructTicTacToeHistory(game, [{ revision: 1, playerId: "user-1", cell: 0 }]);
  throw new Error("history revision gap accepted");
} catch (error) {
  if (!(error instanceof Error) || error.message !== "invalid committed game history") throw error;
}
const moved = validateTicTacToeMove(game, "user-1", 0);
if (moved.board[0] !== "X" || moved.turn !== "user-2")
  throw new Error("legal tic-tac-toe move failed");
try {
  validateTicTacToeMove(moved, "user-1", 1);
  throw new Error("out-of-turn move accepted");
} catch (error) {
  if (!(error instanceof Error) || error.message !== "out-of-turn move") throw error;
}
try {
  validateTicTacToeMove(moved, "user-2", 0);
  throw new Error("occupied tic-tac-toe cell accepted");
} catch (error) {
  if (!(error instanceof Error) || error.message !== "cell occupied") throw error;
}
let winningGame = initializeTicTacToe("dev", "room-win", "user-1", "user-2");
winningGame = validateTicTacToeMove(winningGame, "user-1", 0);
winningGame = validateTicTacToeMove(winningGame, "user-2", 3);
winningGame = validateTicTacToeMove(winningGame, "user-1", 1);
winningGame = validateTicTacToeMove(winningGame, "user-2", 4);
winningGame = validateTicTacToeMove(winningGame, "user-1", 2);
if (ticTacToeWinner(winningGame) !== "user-1") throw new Error("tic-tac-toe win was not detected");
let drawnGame = initializeTicTacToe("dev", "room-draw", "user-1", "user-2");
for (const cell of [0, 1, 2, 4, 3, 5, 7, 6, 8])
  drawnGame = validateTicTacToeMove(drawnGame, drawnGame.turn, cell);
if (!ticTacToeDraw(drawnGame) || ticTacToeWinner(drawnGame))
  throw new Error("tic-tac-toe draw was not detected");
const rulesArtifact = publishTicTacToeRules((source) => `compiled:${source.length}`);
if (!rulesArtifact.startsWith("compiled:")) throw new Error("tic-tac-toe rules were not compiled");
try {
  publishTicTacToeRules(() => "");
  throw new Error("empty rules artifact accepted");
} catch (error) {
  if (!(error instanceof Error) || error.message !== "rules compilation failed") throw error;
}
if (
  authorizedTicTacToeState(game, "dev", "user-1").roomId !== "room-1" ||
  !authorizeTicTacToeCommand(game, "dev", "user-1", "move")
)
  throw new Error("authorized tic-tac-toe access failed");
try {
  authorizedTicTacToeState(game, "prod", "user-1");
  throw new Error("cross-project game state exposed");
} catch (error) {
  if (!(error instanceof Error) || error.message !== "game authorization required") throw error;
}
if (validateTicTacToeCommandPayload({ cell: 4 }).cell !== 4)
  throw new Error("valid command payload rejected");
const pausedGame = pauseTicTacToeTurn(moved);
if (pausedGame.status !== "paused" || pausedGame.board[0] !== "X")
  throw new Error("turn room did not pause at the committed move");
try {
  validateTicTacToeMove(pausedGame, "user-2", 1);
  throw new Error("paused turn accepted a move");
} catch (error) {
  if (!(error instanceof Error) || error.message !== "invalid tic-tac-toe move") throw error;
}
if (resumePausedTicTacToeTurn(pausedGame, "1.0.0", "1.0.0").status !== "active")
  throw new Error("pinned paused room did not resume");
try {
  resumePausedTicTacToeTurn(pausedGame, "2.0.0", "1.0.0");
  throw new Error("room resumed with an unpinned version");
} catch (error) {
  if (!(error instanceof Error) || error.message !== "pinned room version required") throw error;
}
try {
  validateTicTacToeCommandPayload({ cell: 4, extra: true });
  throw new Error("mismatched command payload accepted");
} catch (error) {
  if (!(error instanceof Error) || error.message !== "command payload mismatch") throw error;
}
const coordinator = createTicTacToeCoordinator(game);
const revision = coordinator.revision();
coordinator.move("user-1", 1, revision);
try {
  coordinator.move("user-2", 2, revision);
  throw new Error("concurrent stale move accepted");
} catch (error) {
  if (!(error instanceof Error) || error.message !== "stale game revision") throw error;
}
if (!validateTicTacToeRuleResult({ board: game.board, turn: game.turn, status: "active" }).status)
  throw new Error("valid rule result rejected");
try {
  validateTicTacToeRuleResult({ board: [], turn: "user-1", status: "active", extra: true });
  throw new Error("malformed rule result accepted");
} catch (error) {
  if (!(error instanceof Error) || error.message !== "malformed rule result") throw error;
}
const persistedGame = serializeTicTacToeState(moved);
const restoredGame = reconnectTicTacToe(persistedGame, "dev", "user-2");
if (restoredGame.roomId !== moved.roomId || restoredGame.board[0] !== "X")
  throw new Error("tic-tac-toe state did not reconnect after restart");
const hiddenView = hiddenTicTacToeView(game, "user-1", {
  "user-1": { card: "secret-1" },
  "user-2": { card: "secret-2" },
});
if (
  hiddenView.privateState?.card !== "secret-1" ||
  hiddenTicTacToeView(game, "spectator", { "user-1": { card: "secret-1" } }).privateState !==
    undefined
)
  throw new Error("hidden player state leaked");
