import type { AuthState } from "./auth-state.js";
export function createAuthState(sessionTtlMs = 24 * 60 * 60 * 1000): AuthState { if (sessionTtlMs <= 0) throw new Error("invalid session ttl"); return { accounts: new Map(), sessions: new Map(), rotated: new Set(), sessionTtlMs }; }
