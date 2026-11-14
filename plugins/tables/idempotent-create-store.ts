import { validateProjectId } from "../sdk-typescript/src/index.js";

export class IdempotentCreateStore<T> {
  private readonly results = new Map<string, T>();
  create(projectId: string, key: string, factory: () => T): T {
    validateProjectId(projectId);
    if (!key) throw new Error("idempotency key required");
    const id = `${projectId}:${key}`;
    const existing = this.results.get(id);
    if (existing) return existing;
    const value = factory();
    this.results.set(id, value);
    return value;
  }
}
