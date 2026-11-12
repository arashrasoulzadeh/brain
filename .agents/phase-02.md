# Phase 02 agent guide — co-op arena

Focus: private rooms, FIFO matchmaking, realtime gameplay, reconnect, durable results, dashboard live-session controls, and Eye export.

Keep room state in the room worker. Keep database, compiler, and analytics calls outside the simulation loop. Validate membership, project, plugin version, matchmaking config, input sequence, and room lifecycle before accepting commands.

Use bounded queues, single-use socket tickets, stable result IDs, and asynchronous analytics delivery. Test two browsers first, then reconnect, denial, durable result, Eye delivery, and the eight-player profile.

After every completed checkbox, make a focused commit and push it before taking the next checkbox. After the Phase 02 gate passes, commit and push the phase result.

Before the Phase 02 result commit, update `CHANGELOG.md`, `docs/developers.md`, and `docs/admins.md`.
