# Phase 05 agent guide — lifecycle, UI, and workflows

Focus: promotion, rollback, dependency disable, drain, stop, uninstall, declarative admin views, custom framed views, and project/server workflow extensions.

Use one host-owned lifecycle decision path. Preserve versions pinned by active work. Preview dependency changes with a revision and reject stale applies. Recheck lifecycle and grants at assignment, socket use, job execution, and result commit.

Keep custom UI tokens in the trusted surface. Treat project workflow grants and deployment workflow grants as separate scopes. Plugins never receive publish credentials or the Docker socket.

After every completed checkbox, make a focused commit and push it before taking the next checkbox. After the Phase 05 gate passes, commit and push the phase result.

Before the Phase 05 result commit, update `CHANGELOG.md`, `docs/developers.md`, and `docs/admins.md`.
