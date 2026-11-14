# Playable arena

The arena flow is project-scoped and uses the validated sequence:

1. Create or join a room.
2. Establish a single-use connection ticket.
3. Start the match after membership authorization.
4. Submit bounded player actions.
5. Commit the result with its revision and replay events on reconnect.

The deterministic Tic-tac-toe fixture in `tests/arena/phase02_gate.test.ts`
covers the two-player match, authorized snapshots, reconnect replay, result
validation, and durable completion. Local verification is Docker-free:
`npm run test:ts` and `GOCACHE=/tmp/brain-go-cache go test ./...`.
