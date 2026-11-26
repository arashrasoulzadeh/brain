import { createHash } from "node:crypto";
import { consumeOnce, expiresAt, validateProjectId } from "../sdk-typescript/src/index.js";
import { randomToken, secretsEqual } from "../sdk-typescript/src/tokens.js";

export type AuthorizationCode = {
  projectId: string;
  clientId: string;
  redirectUri: string;
  userId: string;
  expiresAt: number;
  used: boolean;
};

export class AuthorizationCodeIssuer {
  constructor(private readonly codes: Map<string, AuthorizationCode>) {}
  issue(
    projectId: string,
    clientId: string,
    redirectUri: string,
    userId: string,
    now: number,
    ttlSeconds: number,
  ): string {
    validateProjectId(projectId);
    if (!clientId || !redirectUri || !userId || ttlSeconds <= 0)
      throw new Error("invalid authorization code");
    const code = randomToken();
    this.codes.set(code, {
      projectId,
      clientId,
      redirectUri,
      userId,
      expiresAt: expiresAt(now, ttlSeconds * 1000, "invalid authorization code"),
      used: false,
    });
    return code;
  }
}

export class AuthorizationCodeConsumer {
  constructor(private readonly codes: Map<string, AuthorizationCode>) {}
  consume(
    code: string,
    projectId: string,
    clientId: string,
    redirectUri: string,
    now: number,
  ): { userId: string } {
    const record = consumeOnce(this.codes.get(code), now, (record) =>
      record.projectId === projectId && record.clientId === clientId && record.redirectUri === redirectUri,
    );
    if (!record) throw new Error("authorization code unavailable");
    return { userId: record.userId };
  }
}

export function createPKCEChallenge(verifier: string): string {
  return createHash("sha256").update(verifier).digest("base64url");
}

export function verifyPKCE(verifier: string, challenge: string, method: string): boolean {
  if (method !== "S256") return false;
  return secretsEqual(createPKCEChallenge(verifier), challenge);
}

export type OIDCConfig = {
  issuer: string;
  audience: string;
  clientId: string;
  redirectUri: string;
};
export type OIDCTokenClaims = {
  issuer: string;
  audience: string;
  expiresAt: number;
  nonce: string;
  signature: string;
};

export class OIDCConfigReader {
  constructor(private readonly configs: ReadonlyMap<string, OIDCConfig>) {}
  get(projectId: string): OIDCConfig {
    const config = this.configs.get(projectId);
    if (!config) throw new Error("OIDC configuration not found");
    return { ...config };
  }
}

export class OIDCConfigWriter {
  constructor(private readonly configs: Map<string, OIDCConfig>) {}
  set(projectId: string, config: OIDCConfig): void {
    validateProjectId(projectId);
    if (
      !config.audience ||
      !config.clientId ||
      !/^https:\/\//.test(config.issuer) ||
      !/^https:\/\//.test(config.redirectUri)
    )
      throw new Error("invalid OIDC configuration");
    this.configs.set(projectId, { ...config });
  }
}

export type OIDCState = { projectId: string; clientId: string; redirectUri: string };
export class OIDCStateIssuer {
  constructor(private readonly states: Map<string, OIDCState>) {}
  issue(projectId: string, clientId: string, redirectUri: string): string {
    validateProjectId(projectId);
    const state = randomToken();
    this.states.set(state, { projectId, clientId, redirectUri });
    return state;
  }
}

export class OIDCStateConsumer {
  constructor(private readonly states: Map<string, OIDCState>) {}
  consume(state: string, projectId: string, clientId: string, redirectUri: string): void {
    validateProjectId(projectId);
    const record = this.states.get(state);
    if (
      !record ||
      record.projectId !== projectId ||
      record.clientId !== clientId ||
      record.redirectUri !== redirectUri
    )
      throw new Error("callback state mismatch");
    this.states.delete(state);
  }
}

export function validateOIDCToken(
  claims: OIDCTokenClaims,
  config: OIDCConfig,
  expectedNonce: string,
  now: number,
  verifySignature: (signature: string) => boolean,
): void {
  if (
    claims.issuer !== config.issuer ||
    claims.audience !== config.audience ||
    claims.expiresAt <= now ||
    claims.nonce !== expectedNonce ||
    !verifySignature(claims.signature)
  )
    throw new Error("invalid OIDC token");
}
