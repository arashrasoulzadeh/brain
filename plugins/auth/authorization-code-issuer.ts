import { randomBytes } from "node:crypto";
import { validateProjectId } from "../sdk-typescript/src/index.js";
import type { AuthorizationCode } from "./authorization-code-types.js";

export class AuthorizationCodeIssuer {
  constructor(private readonly codes: Map<string, AuthorizationCode>) {}
  issue(projectId: string, clientId: string, redirectUri: string, userId: string, now: number, ttlSeconds: number): string { validateProjectId(projectId); if (!clientId || !redirectUri || !userId || ttlSeconds <= 0) throw new Error("invalid authorization code"); const code = randomBytes(24).toString("base64url"); this.codes.set(code, { projectId, clientId, redirectUri, userId, expiresAt: now + ttlSeconds * 1000, used: false }); return code; }
}
