package com.brain.client

import kotlin.coroutines.Continuation
import kotlin.coroutines.EmptyCoroutineContext
import kotlin.coroutines.startCoroutine
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFailsWith

class BrainClientTest {
    private fun <T> runImmediate(block: suspend () -> T): T {
        var result: Result<T>? = null
        block.startCoroutine(object : Continuation<T> {
            override val context = EmptyCoroutineContext
            override fun resumeWith(value: Result<T>) {
                result = value
            }
        })
        return checkNotNull(result) { "Test transport must complete immediately" }.getOrThrow()
    }

    @Test
    fun projectIdentityIsValidatedBeforeTransport() {
        val transport = AndroidTransport { _, _, _ -> error("must not send") }
        for (invalid in listOf("", " ", "../dev", "dev/project", "-dev", "_dev", "dev\n", "é")) {
            assertFailsWith<IllegalArgumentException>(invalid) { BrainClient(invalid, transport) }
        }
        for (valid in listOf("dev", "A", "9", "dev-01_test")) {
            BrainClient(valid, transport)
        }
    }

    @Test
    fun signInSendsExactProjectAndCredentials() {
        val response = SignInResponse("session-token", "alice")
        val client = BrainClient("project-01", AndroidTransport { operation, project, payload ->
            assertEquals("auth.signin", operation)
            assertEquals("project-01", project)
            assertEquals(mapOf("username" to "alice", "password" to "test-password"), payload)
            response
        })
        assertEquals(response, runImmediate { client.signIn("alice", "test-password") })
    }

    @Test
    fun malformedSignInResponseIsRejected() {
        val client = BrainClient("dev", AndroidTransport { _, _, _ -> mapOf("token" to "untyped") })
        assertFailsWith<IllegalArgumentException> { runImmediate { client.signIn("alice", "test-password") } }
    }

    @Test
    fun callbackRequiresHttpsExpectedOriginAndBothParameters() {
        val prefix = "https://app.example.test/auth/callback?"
        assertEquals(AuthCallback("abc", "xyz"), parseAuthCallback(prefix + "code=abc&state=xyz"))
        assertEquals(AuthCallback("a=b", "xyz"), parseAuthCallback(prefix + "state=xyz&code=a=b"))
        for (invalid in listOf(
            "http://app.example.test/auth/callback?code=abc&state=xyz",
            "https://other.example.test/auth/callback?code=abc&state=xyz",
            "https://app.example.test/wrong?code=abc&state=xyz",
            prefix + "code&state=xyz",
            prefix + "state=xyz",
            prefix + "code=&state=xyz",
            prefix + "code=abc",
            prefix + "code=abc&state="
        )) {
            assertFailsWith<IllegalArgumentException>(invalid) { parseAuthCallback(invalid) }
        }
    }
}
