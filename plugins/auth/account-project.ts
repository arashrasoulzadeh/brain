import type { Account, AuthState } from "./auth-state.js";
export function accountProject(state: AuthState, projectId: string): Map<string, Account> { const project = state.accounts.get(projectId) ?? new Map<string, Account>(); state.accounts.set(projectId, project); return project; }
