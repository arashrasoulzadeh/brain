import type { SchemaChangePlan } from "./index.js";
import type { SchemaAuditState } from "./schema-audit-state.js";
export class SchemaPublicationRecorder {
  constructor(private readonly state: SchemaAuditState) {}
  record(projectId: string, actorId: string, plan: SchemaChangePlan, publishedAt: number): void {
    if (!actorId || plan.projectId !== projectId || !Number.isFinite(publishedAt))
      throw new Error("invalid schema publication audit");
    this.state.entries.push({
      ...plan,
      actorId,
      publishedAt,
      change:
        plan.change.kind === "create_index"
          ? { ...plan.change, columns: [...plan.change.columns] }
          : { ...plan.change },
    });
  }
}
