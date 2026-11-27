# Live sessions

Route: `/{project}/ui/live-sessions`

A manifest alias rendered from eligible contributions. Rooms exposes details and management (create, join, reconnect, result); Matchmaking contributes ticket inventory/status/queue/assignment/cancel; compatible Arena and Invites contributions can appear as independently authorized extensions.

Use the selected active project and the resource inventory to locate a room or ticket. Read-only queries can refresh after declared observations; writes require explicit confirmation and are never automatically replayed. Verify project and resource identity before changing a room or match.

This console is for trusted operators. It is not a public realtime socket server or a game client.

See [Rooms](../plugin-guides/rooms.md), [Arena](../plugin-guides/arena.md), [Matchmaking](../plugin-guides/matchmaking.md), [Invites](../plugin-guides/invites.md), and [extension contracts](extensions.md).
