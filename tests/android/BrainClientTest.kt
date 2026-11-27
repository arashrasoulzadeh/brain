package com.brain.client

import kotlin.coroutines.Continuation
import kotlin.coroutines.EmptyCoroutineContext
import kotlin.coroutines.startCoroutine
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFailsWith
import kotlin.test.assertFalse
import kotlin.test.assertTrue

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

    @Test
    fun clientResponsesExposeStableValueSemantics() {
        val session = SignInResponse("session-token", "alice")
        assertEquals("session-token", session.token)
        assertEquals("alice", session.username)
        assertTrue(session.equals(session))
        assertFalse(session.equals(null))
        assertFalse(session.equals("session-token"))
        assertFalse(session == SignInResponse("different", "alice"))
        assertFalse(session == SignInResponse("session-token", "different"))
        val (token, username) = session
        assertEquals("session-token", token)
        assertEquals("alice", username)
        assertEquals(session, session.copy())
        assertEquals(SignInResponse("new-token", "alice"), session.copy(token = "new-token"))
        assertEquals(SignInResponse("session-token", "bob"), session.copy(username = "bob"))
        assertEquals(session.hashCode(), session.copy().hashCode())
        assertEquals("SignInResponse(token=session-token, username=alice)", session.toString())
        val callback = AuthCallback("abc", "xyz")
        assertEquals("abc", callback.code)
        assertEquals("xyz", callback.state)
        assertTrue(callback.equals(callback))
        assertFalse(callback.equals(null))
        assertFalse(callback.equals("abc"))
        assertFalse(callback == AuthCallback("different", "xyz"))
        assertFalse(callback == AuthCallback("abc", "different"))
        val (code, state) = callback
        assertEquals("abc", code)
        assertEquals("xyz", state)
        assertEquals(callback, callback.copy())
        assertEquals(AuthCallback("new", "xyz"), callback.copy(code = "new"))
        assertEquals(AuthCallback("abc", "new"), callback.copy(state = "new"))
        assertEquals(callback.hashCode(), callback.copy().hashCode())
        assertEquals("AuthCallback(code=abc, state=xyz)", callback.toString())
    }
}
