# Playable arena

The arena flow is project-scoped and uses the validated sequence:

1. Create or join a room.
2. Establish a single-use connection ticket.
3. Start the match after membership authorization.
4. Submit bounded player actions.
5. Commit the result with its revision and replay events on reconnect.

Game-specific Tic-Tac-Toe rules live in the Games plugin; generic turn and revision helpers are exposed by PvP. The room plugin remains responsible for membership and transport-independent lifecycle. The Tic-Tac-Toe tests are in `tests/games/games.test.ts` and `tests/functions/phase04_tictactoe_gate.test.ts`; the cooperative arena gate remains in `tests/arena/phase02_gate.test.ts`. Local verification is Docker-free:
`npm run test:ts` and `GOCACHE=/tmp/brain-go-cache go test ./...`.
