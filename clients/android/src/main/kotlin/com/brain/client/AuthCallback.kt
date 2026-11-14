package com.brain.client

data class AuthCallback(val code: String, val state: String)

fun parseAuthCallback(uri: String): AuthCallback {
    val prefix = "https://app.example.test/auth/callback?"
    require(uri.startsWith(prefix)) { "invalid auth callback" }
    val values = uri.removePrefix(prefix).split('&').associate {
        val parts = it.split('=', limit = 2)
        require(parts.size == 2) { "invalid auth callback" }
        parts[0] to parts[1]
    }
    val code = values["code"]
    val state = values["state"]
    require(!code.isNullOrEmpty() && !state.isNullOrEmpty()) { "invalid auth callback" }
    return AuthCallback(code, state)
}
