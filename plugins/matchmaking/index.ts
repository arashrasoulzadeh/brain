import { request } from "../sdk-typescript/src/index.js";

export type MatchmakingConfig = { projectId: string; scope: string; matchSize: number; ticketTtlSeconds: number; enabled: boolean };

export function validateConfig(config: MatchmakingConfig): void {
  if (!/^[a-zA-Z0-9][a-zA-Z0-9_-]*$/.test(config.projectId) || !/^[a-zA-Z][a-zA-Z0-9_-]*$/.test(config.scope)) throw new Error("invalid matchmaking scope");
  if (config.matchSize < 1 || config.matchSize > 64 || config.ticketTtlSeconds < 1 || config.ticketTtlSeconds > 3600) throw new Error("invalid matchmaking config");
}

export class MatchmakingConfigs {
  private readonly configs = new Map<string, MatchmakingConfig>();
  set(config: MatchmakingConfig): MatchmakingConfig {
    validateConfig(config);
    const key = `${config.projectId}:${config.scope}`;
    this.configs.set(key, { ...config });
    return { ...config };
  }
  get(projectId: string, scope: string): MatchmakingConfig {
    const config = this.configs.get(`${projectId}:${scope}`);
    if (!config) throw new Error("matchmaking config not found");
    return { ...config };
  }
  isEnabled(projectId: string, scope: string): boolean { return this.configs.get(`${projectId}:${scope}`)?.enabled ?? false; }
}

export type MatchmakingTicket = { id: string; projectId: string; scope: string; playerId: string; createdAt: number; expiresAt: number };

export class MatchmakingQueue {
  private sequence = 0;
  private readonly queues = new Map<string, MatchmakingTicket[]>();
  constructor(private readonly configs?: MatchmakingConfigs) {}
  enqueue(projectId: string, scope: string, playerId: string, now = Date.now()): MatchmakingTicket {
    if (!projectId || !scope || !playerId) throw new Error("invalid matchmaking ticket");
    if (this.configs && !this.configs.isEnabled(projectId, scope)) throw new Error("matchmaking scope disabled");
    const key = `${projectId}:${scope}`;
    if ((this.queues.get(key) ?? []).some((ticket) => ticket.playerId === playerId)) throw new Error("active ticket already exists");
    const ttl = this.configs?.get(projectId, scope).ticketTtlSeconds ?? 60;
    const ticket = { id: `${key}:ticket-${++this.sequence}`, projectId, scope, playerId, createdAt: now, expiresAt: now + ttl * 1000 };
    this.queues.set(key, [...(this.queues.get(key) ?? []), ticket]);
    return { ...ticket };
  }
  peek(projectId: string, scope: string): readonly MatchmakingTicket[] { return [...(this.queues.get(`${projectId}:${scope}`) ?? [])]; }
  match(projectId: string, scope: string, matchSize: number): MatchmakingTicket[] {
    const key = `${projectId}:${scope}`;
    const queue = this.queues.get(key) ?? [];
    if (matchSize < 1 || queue.length < matchSize) return [];
    const matched = queue.slice(0, matchSize);
    this.queues.set(key, queue.slice(matchSize));
    return matched.map((ticket) => ({ ...ticket }));
  }
  expire(now = Date.now()): MatchmakingTicket[] {
    const expired: MatchmakingTicket[] = [];
    for (const [key, queue] of this.queues) {
      const active = queue.filter((ticket) => { if (ticket.expiresAt <= now) { expired.push({ ...ticket }); return false; } return true; });
      this.queues.set(key, active);
    }
    return expired;
  }
}

export function matchmakingRequest(projectId: string, pluginId: string, version: string, operation: "matchmaking.queue" | "matchmaking.cancel" | "matchmaking.status" | "matchmaking.assign", config: string, payload: unknown = {}) {
  return request(`matchmaking-${operation}-${projectId}-${config}`, { protocol: 1, projectId, pluginId, version, operation, payload: { config, ...(payload as Record<string, unknown>) } });
}

