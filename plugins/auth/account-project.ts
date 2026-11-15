import type { Account, AuthState } from "./auth-state.js";
import { validateProjectId } from "../sdk-typescript/src/index.js";
export function accountProject(state: AuthState, projectId: string): Map<string, Account> {
  validateProjectId(projectId);
  const project = state.accounts.get(projectId) ?? new Map<string, Account>();
  state.accounts.set(projectId, project);
  return project;
}
