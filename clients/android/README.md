# Android client

The Kotlin client uses a project-bound coroutine transport. App Links must
resolve to `https://app.example.test/auth/callback` and include both `code` and
`state`; insecure or incomplete callbacks are rejected. Refresh credentials
must be implemented behind Android Keystore-backed storage before production
release.
