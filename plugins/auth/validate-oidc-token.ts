import type { OIDCConfig, OIDCTokenClaims } from "./oidc-types.js";
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
