import { validateProjectId } from "../sdk-typescript/src/index.js";
export function validateAccount(projectId: string, username: string, password: string): void {
  validateProjectId(projectId);
  if (!username || password.length < 8) throw new Error("invalid signup");
}
