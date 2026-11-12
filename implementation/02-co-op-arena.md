# Phase 02 — Co-op arena

[Phase checklist](../phases.md) · [Architecture plan](../plan.md)

Use the project, plugin, permission, and analytics foundations from Phase 01. Deliver a browser-playable arena on local Compose before expanding the app features. Start each slice with the failing test, make it pass with the smallest change, then refactor while the tests stay green.

## 1. Open the arena as a guest

- [ ] Write a browser test for creating a guest, refreshing the page, and signing out. Assert that another project rejects the session.
- [ ] Connect the arena page to the guest plugin and shared session endpoints from Phase 01.
- [ ] Write a handshake test for an expired ticket, a reused ticket, and a disallowed browser origin.
- [ ] Add the short-lived gameplay socket ticket flow and return a recoverable error to the browser when the connection fails.

## 2. Start a scoped runtime worker

- [ ] Write a host-to-worker contract test for a granted arena call and denials for the wrong project, plugin version, room, or matchmaking config.
- [ ] Register the bundled Go arena plugin and start it under the isolated worker supervisor.
- [ ] Connect the host and worker over authenticated private IPC. Keep the worker behind the shared permission broker.
- [ ] Verify that disabling or revoking the arena capability closes affected access and prevents late writes. Use the lifecycle from Phase 01.

## 3. Create and join a private room

- [ ] Write tests for creating a room, using an expiring invite, racing for the final seat, and denying access to a nonmember.
- [ ] Add project-scoped room creation and admission, with a server-owned room ID, a pinned rules version, and support for two to eight players.
- [ ] Add the invite plugin and browser controls for creating, sharing, and joining a private room.
- [ ] Check membership on both room lookup and socket admission through the same policy boundary.

## 4. Match two players in FIFO order

- [ ] Write transaction tests for duplicate enqueue, expired tickets, FIFO selection within a config, and cancellation racing with assignment.
- [ ] Add a two-player arena matchmaking config and group tickets by project, config revision, and compatibility.
- [ ] Reserve seats, create the provisioning room, and mark tickets assigned in one transaction. Recheck effective enablement before commit.
- [ ] Test a restart between reservation and room activation. Recover provisioning or expire the reservation with a reason the client can act on.
- [ ] Send assignments on the private gameplay control channel and let the browser recover the assignment through HTTPS.

## 5. Play one authoritative match

- [ ] Write fixed-clock tests for legal movement, a forged position, replayed input, and completion of the shared objective.
- [ ] Implement the arena rules behind the room runtime contract. Use one room owner, a 20 Hz initial simulation rate, and up to 10 snapshots per second.
- [ ] Sequence inputs per player. Bound queues, coalesce continuous movement, deduplicate discrete actions, and drop superseded snapshots.
- [ ] Put the server tick and last resolved input sequence in each authorized snapshot. Render and interpolate that state in the browser.
- [ ] Test a slow consumer and an overloaded tick. Keep database, compiler, and Eye calls out of the simulation loop.

## 6. Reconnect and record the result

- [ ] Write tests for reconnect inside the grace period, replacement of the old socket, duplicate result submission, and loss of the worker.
- [ ] Return an authoritative snapshot and input cursor after reconnect. Confirm that the player can finish the same match.
- [ ] Commit a single server-computed result per match before reporting completion. Recheck ownership and the lifecycle revision at commit.
- [ ] Mark a match interrupted when its worker is lost. Verify that the status API reports the interruption after restart.

## 7. Show a live session in the dashboard

- [ ] Add an operator test that configures the arena, launches a test room, and sees membership, queue status, and the recorded result. Reject the same actions from a player session.
- [ ] Add Go-rendered live-session controls and status views over the existing host operations. Reuse the arena and matchmaking settings for test and player rooms.

## 8. Send the match outcome to Eye

- [ ] Write an exporter contract test for the Eye tenant key and session, the event allowlist, stable retry IDs, aggregate rejection, and a disabled scope.
- [ ] Enqueue the outcome through the analytics event contract and send it asynchronously through the bounded Eye spool.
- [ ] Verify that one completed match produces the expected event in an Eye test tenant. Confirm that an Eye outage leaves the arena available.

## 9. Run the local pilot

- [ ] Start from fresh local Compose state and run the two-browser journey from sign-in through matchmaking, reconnect, result, and Eye delivery.
- [ ] Run an eight-player room with simulated latency and packet loss. Record input-to-snapshot latency, tick delays, memory, CPU, and outbound traffic.
- [ ] Refactor shared room and ticket rules into their owning components. Keep the game rules independent of transport and persistence.

## Exit gate

- [ ] A fresh local stack supports a private match and a queued co-op match. Dashboard controls, reconnect, one committed result, scoped plugin permissions, and Eye delivery pass end to end.
- [ ] Record early performance measurements as a baseline for the 4-vCPU, 8-GB VPS. Set the room limit after the mixed-load release test.
