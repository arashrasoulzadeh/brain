# Android and Kotlin client

The [Kotlin client](src/main/kotlin/com/brain/client/BrainClient.kt) is a coroutine-based transport contract. Implement `AndroidTransport.send(operation, projectId, payload)` in your trusted application integration, then construct `BrainClient("dev", transport)`. `signIn(username, password)` invokes `auth.signin` and requires a typed `SignInResponse(token, username)`. The client validates project identifiers; it does not authenticate project membership by itself.

The current production host exposes published functions and operator HTTP routes, not a generic user RPC endpoint. See the [HTTP reference](../../docs/http-api.md). Do not put an operator session in an Android app.

App Links must resolve to `https://app.example.test/auth/callback` and include both `code` and `state`; insecure or incomplete callbacks are rejected. Refresh credentials must be implemented behind Android Keystore-backed storage before production release. The package is a Kotlin/JVM client and test fixture, not a shipped Android application or an integrated OAuth provider.

Build and test using Gradle from `clients/android`; see [coverage CI](../../.github/workflows/coverage.yml) for the pinned JDK/Gradle invocation and JaCoCo gate.
