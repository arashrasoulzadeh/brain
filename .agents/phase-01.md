# Phase 01 agent guide — projects and plugin host

Focus: project isolation, operator access, plugin manifests, grants, worker supervision, and the first analytics event.

Every plugin call must carry project, actor or trigger identity, installation, resource scope, and lifecycle revision. Validate all of them in the host, not in workers. Keep bundled and external workers behind the same contract tests.

Do not implement arena-specific behavior here except the guest session and event contract needed by Phase 02. Test two projects in every isolation fixture.

After every completed checkbox, make a focused commit and push it before taking the next checkbox. After the Phase 01 gate passes, commit and push the phase result.
