export type SubscriptionTicket = { ticket: string; projectId: string; channel: string; operation: string; expiresAt: number };
export type SubscriptionEvent<T> = { sequence: number; value: T };
