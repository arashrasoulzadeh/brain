import type { Invite } from "./types.js";

export class InviteAccepter {
  constructor(private readonly invites: Map<string, Invite>) {}
  accept(inviteId: string, projectId: string, playerId: string, now = Date.now()): Invite { const invite = this.invites.get(inviteId); if (!invite || invite.projectId !== projectId || invite.targetPlayerId !== playerId || invite.used || invite.expiresAt <= now) throw new Error("invite unavailable"); invite.used = true; return { ...invite }; }
}
