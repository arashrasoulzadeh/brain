import type { FormAudit } from "./types.js";

export class AuditedFormActions {
  private readonly entries: FormAudit[] = [];
  record(projectId: string, actorId: string, action: string, occurredAt: number): void { if (!projectId || !actorId || !action || !Number.isFinite(occurredAt)) throw new Error("invalid form audit"); this.entries.push({ projectId, actorId, action, occurredAt }); }
  list(projectId: string): FormAudit[] { return this.entries.filter((entry) => entry.projectId === projectId).map((entry) => ({ ...entry })); }
}
