import { validateProjectId } from "../sdk-typescript/src/index.js";

export class QuotaReader {
  constructor(private readonly reserved: ReadonlyMap<string, number>) {}
  reservedFor(projectId: string): number {
    validateProjectId(projectId);
    return this.reserved.get(projectId) ?? 0;
  }
}
