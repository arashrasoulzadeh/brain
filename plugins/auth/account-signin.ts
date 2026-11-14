import { verify as argon2Verify } from "@node-rs/argon2";
import { randomBytes } from "node:crypto";
import { validateProjectId } from "../sdk-typescript/src/index.js";
import { sessionTokenDigest } from "./session-token-digest.js";
import type { AuthState } from "./auth-state.js";
export class AccountSignin {
  constructor(private readonly state: AuthState) {}
  async signin(projectId: string, username: string, password: string, now = Date.now()) {
    validateProjectId(projectId);
    const account = this.state.accounts.get(projectId)?.get(username);
    if (!account || !(await argon2Verify(account.passwordHash, password)))
      throw new Error("invalid credentials");
    const token = randomBytes(32).toString("hex");
    this.state.sessions.set(sessionTokenDigest(token), {
      projectId,
      username,
      expiresAt: now + this.state.sessionTtlMs,
    });
    return { token, username };
  }
}
