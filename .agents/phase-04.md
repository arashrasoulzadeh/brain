# Phase 04 agent guide — functions and turns

Focus: offline Go compilation, WASM execution, publication, durable triggers, workflows, and turn-based rooms.

Compiler and runner are separate trust boundaries. Pin toolchain, SDK, dependencies, artifact digest, and source bundle. Disable network, credentials, filesystem access, CGO, downloads, and user build commands. Bound CPU, memory, processes, scratch space, diagnostics, input, output, and wall time.

Load only declared inputs. Commit validated effects, result, job completion, and idempotency state together. Keep room rules pure and outside database locks. Recheck membership, version, lifecycle, and grants before committing a turn.

After every completed checkbox, make a focused commit and push it before taking the next checkbox. After the Phase 04 gate passes, commit and push the phase result.

Before the Phase 04 result commit, update `CHANGELOG.md`, `docs/developers.md`, and `docs/admins.md`.
