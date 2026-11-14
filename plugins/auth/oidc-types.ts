export type OIDCConfig = { issuer: string; audience: string; clientId: string; redirectUri: string };
export type OIDCTokenClaims = { issuer: string; audience: string; expiresAt: number; nonce: string; signature: string };
