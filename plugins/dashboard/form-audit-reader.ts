import type { FormAudit } from "./types.js";

export class FormAuditReader {
  constructor(private readonly entries: readonly FormAudit[]) {}
  list(projectId: string): FormAudit[] {
    return this.entries
      .filter((entry) => entry.projectId === projectId)
      .map((entry) => ({ ...entry }));
  }
}
