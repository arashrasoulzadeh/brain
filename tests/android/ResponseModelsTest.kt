package com.brain.client

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class ResponseModelsTest {
    @Test
    fun signInResponseExposesFieldsAndValueEquality() {
        val original = SignInResponse("token", "alice")
        assertEquals("token", original.token)
        assertEquals("alice", original.username)
        assertEquals("token", original.component1())
        assertEquals("alice", original.component2())
        assertTrue(original.equals(original))
        assertTrue(original.equals(SignInResponse("token", "alice")))
        assertFalse(original.equals(null))
        assertFalse(original.equals("token"))
        assertFalse(original.equals(SignInResponse("other", "alice")))
        assertFalse(original.equals(SignInResponse("token", "bob")))
        assertEquals(SignInResponse("new", "bob"), original.copy("new", "bob"))
        assertEquals(original, original.copy())
        assertEquals(SignInResponse("new", "alice"), original.copy(token = "new"))
        assertEquals(SignInResponse("token", "bob"), original.copy(username = "bob"))
        assertEquals(original.hashCode(), original.copy().hashCode())
        assertEquals("SignInResponse(token=token, username=alice)", original.toString())
    }

    @Test
    fun authCallbackExposesFieldsAndValueEquality() {
        val original = AuthCallback("code", "state")
        assertEquals("code", original.code)
        assertEquals("state", original.state)
        assertEquals("code", original.component1())
        assertEquals("state", original.component2())
        assertTrue(original.equals(original))
        assertTrue(original.equals(AuthCallback("code", "state")))
        assertFalse(original.equals(null))
        assertFalse(original.equals("code"))
        assertFalse(original.equals(AuthCallback("other", "state")))
        assertFalse(original.equals(AuthCallback("code", "other")))
        assertEquals(AuthCallback("new", "new-state"), original.copy("new", "new-state"))
        assertEquals(original, original.copy())
        assertEquals(AuthCallback("new", "state"), original.copy(code = "new"))
        assertEquals(AuthCallback("code", "new-state"), original.copy(state = "new-state"))
        assertEquals(original.hashCode(), original.copy().hashCode())
        assertEquals("AuthCallback(code=code, state=state)", original.toString())
    }
}
