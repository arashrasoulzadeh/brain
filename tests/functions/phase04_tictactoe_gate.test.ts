import { initializeTicTacToe, ticTacToeWinner, validateTicTacToeMove } from "../../plugins/rooms/index.js";

let game = initializeTicTacToe("dev", "gate-room", "player-a", "player-b");
for (const [player, cell] of [["player-a", 0], ["player-b", 3], ["player-a", 1], ["player-b", 4], ["player-a", 2]] as const) game = validateTicTacToeMove(game, player, cell);
if (ticTacToeWinner(game) !== "player-a" || game.status !== "active") throw new Error("two-client Tic-tac-toe match did not complete");
