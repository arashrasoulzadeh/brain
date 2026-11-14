export type AuthSettings = {
  passwordEnabled: boolean;
  oidcProviders: readonly string[];
  emailRecoveryEnabled: boolean;
};
export type SignupSettings = {
  enabled: boolean;
  allowedDomains: readonly string[];
  requireInvite: boolean;
};
export type FormAudit = { projectId: string; actorId: string; action: string; occurredAt: number };
