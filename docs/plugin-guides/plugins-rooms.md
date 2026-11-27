# Rooms

## Developer

Create with `rooms.create` payload `{"id":"r1"}`; join with `rooms.join` payload `{"id":"r1","member":"alice"}`. Reconnect uses the same identity fields. Read or commit a result using `rooms.result` and the room ID. A successful join or reconnect emits the declared connected observation. The member field is not an authentication credential.

## Administrator

Grant the room operations by resource and verify trusted user membership before any public integration. Test persisted room state, reconnect, result commit, and observation delivery in the same project. Drain active rooms before disabling or promoting; consult [playable arena](../playable-arena.md) and [deployment](../deployment-runbook.md).
