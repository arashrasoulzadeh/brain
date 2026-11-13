export type MatchmakingConfig = { projectId: string; scope: string; matchSize: number; ticketTtlSeconds: number; enabled: boolean };
export type MatchmakingTicket = { id: string; projectId: string; scope: string; playerId: string; createdAt: number; expiresAt: number };
export type AssignmentDecision = { id: string; projectId: string; scope: string; ticketIds: readonly string[]; createdAt: number };
