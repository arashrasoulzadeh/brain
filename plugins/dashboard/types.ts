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
export type FunctionSummary = { id: string; version: number; status: string };
export type WorkflowSummary = { id: string; status: string; lastRunAt?: number };
export type LiveSessionSummary = {
  id: string;
  kind: string;
  status: string;
  participantCount: number;
};
