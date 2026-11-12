# Phase 00 agent guide — foundation

Focus: minimal Go host, PostgreSQL, migrations, Compose, and CI.

Start with the first unchecked Phase 00 item in `roadmap.md`. Keep the host small: command setup, configuration, HTTP package, database package, and health endpoints.

Verify in this order: focused Go test, package tests, migration test, Compose smoke test, then CI configuration checks.

After every completed checkbox, make a focused commit and push it before taking the next checkbox. After the Phase 00 gate passes, commit and push the phase result.

Before the Phase 00 result commit, update `CHANGELOG.md`, `docs/developers.md`, and `docs/admins.md`.

Do not add projects, plugins, auth, or gameplay yet. Runtime credentials must not run migrations or DDL.
