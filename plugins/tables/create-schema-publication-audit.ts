import type { SchemaChange, SchemaChangePlan, SchemaPublication } from "./index.js";

export function createSchemaPublicationAudit() {
  const entries: SchemaPublication[] = [];
  return {
    record(projectId: string, actorId: string, plan: SchemaChangePlan, publishedAt: number): void {
      if (!actorId || plan.projectId !== projectId || !Number.isFinite(publishedAt))
        throw new Error("invalid schema publication audit");
      entries.push({ ...plan, actorId, publishedAt, change: copyChange(plan.change) });
    },
    list(projectId: string) {
      return entries.filter((entry) => entry.projectId === projectId).map((entry) => ({
        ...entry,
        change: copyChange(entry.change),
      }));
    },
  };
}

function copyChange(change: SchemaChange): SchemaChange {
  return change.kind === "create_index" ? { ...change, columns: [...change.columns] } : { ...change };
}
