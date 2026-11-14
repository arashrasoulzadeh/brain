package com.brain.client

data class SignInResponse(val token: String, val username: String)

fun interface AndroidTransport {
    suspend fun send(operation: String, projectId: String, payload: Map<String, String>): Any
}

class BrainClient(private val projectId: String, private val transport: AndroidTransport) {
    init {
        require(projectId.isNotBlank()) { "client project required" }
    }

    suspend fun signIn(username: String, password: String): SignInResponse {
        val result = transport.send("auth.signin", projectId, mapOf("username" to username, "password" to password))
        require(result is SignInResponse) { "invalid sign-in response" }
        return result
    }
}
