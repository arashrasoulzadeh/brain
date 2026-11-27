# Games

Game-specific rules. [Authoritative manifest](../../plugins/games/plugin.yaml).

## Developer

Import the Tic-Tac-Toe initialization, move validation, winner/draw, recovery and replay helpers from Games. Reuse the shared PvP primitives instead of copying turn machinery. Game-specific logic does not belong in Rooms.

## Administrator

Install and grant games.modes to list available modes. The discovery call is read-only and has no game session side effect. Keep game versions compatible with persisted state in your application.

## Example payloads

games.modes: {} returns projectId and modes including tic-tac-toe. For rules and replay see [playable arena](../playable-arena.md).

Payload snippets are arguments to the declared operation, not standalone HTTP routes. Use the [HTTP invocation contract](../http-api.md) and exact manifest version/capability. Host-owned context, internal snapshots and grants must not be supplied by public callers.

## Limitations

The discovery route is not a public Tic-Tac-Toe command server; local TypeScript mechanics require a trusted application integration.

The Game modes workspace also exposes read-only `games.mode.describe` with `{"mode":"tic-tac-toe"}` for the built-in rules, board dimensions, player count and move contract. It uses the existing games.modes capability; it is not a persisted game-instance manager.
