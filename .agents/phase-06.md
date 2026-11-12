# Phase 06 agent guide — clients and release

Focus: browser, Flutter, Android, Zarvan fixtures, multi-architecture images, manual Compose deployment, backup/restore, and capacity measurement.

Keep all clients aligned to shared HTTP and WebSocket fixtures. Retries must be safe reads or use explicit idempotency keys. Native clients keep refresh credentials in platform secure storage.

Release artifacts are keyed by the full commit SHA and pinned image digests. Pull and validate images before draining. Preserve the last healthy set. Test migration failure, health failure, rollback, restore, both architectures, and the measured 4-vCPU/8-GB profile.

After every completed checkbox, make a focused commit and push it before taking the next checkbox. After the Phase 06 gate passes, commit and push the phase result.
