# Arena

## Developer

Start with `arena.start` payload `{"roomId":"r1","players":["a","b"]}`. Submit actions with `arena.input` using the room ID, player ID, monotonically increasing integer sequence and payload. Read `arena.snapshot` and finish with `arena.result`, including winner IDs and a valid ISO completion time. Duplicate players, stale sequences and unknown winners are rejected.

## Administrator

Grant the four arena capabilities. Rooms and Matchmaking are optional integrations; the host must authorize actual membership before exposing these operator routes to users. Arena worker state is in memory; do not assume recovery across worker restarts. See [playable arena](../playable-arena.md) for the broader flow.
