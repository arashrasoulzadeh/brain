# Live sessions

Route: `/{project}/ui/live-sessions`

A compatibility manifest alias rendered from eligible contributions. Rooms and Matchmaking now have dedicated sidebar workspaces at `/{project}/ui/rooms` and `/{project}/ui/matchmaking`; Arena and Invitations have their own destinations too. Rooms exposes details and management (create, join, reconnect, result); Matchmaking contributes ticket inventory/status/queue/assignment/cancel; compatible Arena and Invites contributions can appear as independently authorized extensions.

Use the selected active project and the resource inventory to locate a room or ticket. Read-only queries can refresh after declared observations; writes require explicit confirmation and are never automatically replayed. Verify project and resource identity before changing a room or match.

This console is for trusted operators. It is not a public realtime socket server or a game client.

See [Rooms](../plugin-guides/rooms.md), [Arena](../plugin-guides/arena.md), [Matchmaking](../plugin-guides/matchmaking.md), [Invites](../plugin-guides/invites.md), and [extension contracts](extensions.md).

Authorized room inventories show stored membership counts and open/closed state; ticket inventories show persisted assignment status. Room management supports explicit membership removal and closure, subject to its new resource-scoped grant. This does not manage active sockets or a distributed scheduler.

Arena matches expose persisted match status and player count, and the confirmed interruption operation terminates an active match without awarding a result. The stored interruption reason and time remain visible in match inspection.
