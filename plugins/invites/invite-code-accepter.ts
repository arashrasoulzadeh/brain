import type { Invite } from "./types.js";
import { InviteAccepter } from "./invite-accepter.js";

export class InviteCodeAccepter {
  private readonly accepter: InviteAccepter;
  constructor(private readonly invites: Map<string, Invite>) { this.accepter = new InviteAccepter(invites); }
  acceptCode(code: string, projectId: string, playerId: string, now = Date.now()): Invite { const invite = [...this.invites.values()].find((candidate) => candidate.code === code); if (!invite) throw new Error("invite unavailable"); return this.accepter.accept(invite.id, projectId, playerId, now); }
}
