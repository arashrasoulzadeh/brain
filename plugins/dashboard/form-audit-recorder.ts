import type { FormAudit } from "./types.js";

export class FormAuditRecorder {
  constructor(private readonly entries: FormAudit[]) {}
  record(projectId: string, actorId: string, action: string, occurredAt: number): void { if (!projectId || !actorId || !action || !Number.isFinite(occurredAt)) throw new Error("invalid form audit"); this.entries.push({ projectId, actorId, action, occurredAt }); }
}
