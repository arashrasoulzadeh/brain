# Phase 03 agent guide — app data and auth

Focus: project-scoped identity, sessions, invitations, OIDC, OTP delivery, PostgreSQL table exposure, files, subscriptions, and dashboard controls.

Use real PostgreSQL roles and transactions for isolation tests. Bind caller context transaction-locally. Default to deny. Validate table identifiers and filters. Use record versions and idempotency keys for retried writes.

Never log codes, tokens, credentials, or raw sensitive payloads. Keep fake email/SMS delivery test-only. Recheck authorization during subscription delivery, file access, and destructive dashboard actions.

After every completed checkbox, make a focused commit and push it before taking the next checkbox. After the Phase 03 gate passes, commit and push the phase result.
