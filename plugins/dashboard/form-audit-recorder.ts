import type { FormAudit } from "./types.js";
import { validateProjectId } from "../sdk-typescript/src/index.js";

export class FormAuditRecorder {
  constructor(private readonly entries: FormAudit[]) {}
  record(projectId: string, actorId: string, action: string, occurredAt: number): void {
    try {
      validateProjectId(projectId);
    } catch {
      throw new Error("invalid form audit");
    }
    if (!actorId || !action || !Number.isFinite(occurredAt)) throw new Error("invalid form audit");
    this.entries.push({ projectId, actorId, action, occurredAt });
  }
}
