import { validateProjectId, validateIdentifier } from "../sdk-typescript/src/index.js";

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

export function paginate<T>(
  items: readonly T[],
  cursor: string | undefined,
  limit: number,
): { items: T[]; nextCursor?: string } {
  if (!Number.isInteger(limit) || limit < 1 || limit > 100) throw new Error("invalid page limit");
  const start = cursor ? Number(Buffer.from(cursor, "base64url").toString("utf8")) : 0;
  if (!Number.isInteger(start) || start < 0 || start > items.length)
    throw new Error("invalid page cursor");
  const page = [...items.slice(start, start + limit)];
  return {
    items: page,
    ...(start + page.length < items.length
      ? { nextCursor: Buffer.from(String(start + page.length)).toString("base64url") }
      : {}),
  };
}

export function validateFilter(
  filter: Record<string, unknown>,
  approvedColumns: readonly string[],
): void {
  for (const [column, expression] of Object.entries(filter)) {
    if (
      !approvedColumns.includes(column) ||
      typeof expression !== "object" ||
      expression === null ||
      !Object.keys(expression as object).every((operator) =>
        ["eq", "lt", "gt", "in"].includes(operator),
      )
    )
      throw new Error("unapproved filter");
  }
}

export function validateValue(column: string, value: unknown): void {
  validateIdentifier(column);
  if (
    value === undefined ||
    typeof value === "function" ||
    (typeof value === "string" && value.length > 4096)
  )
    throw new Error("invalid value");
}

export function createVersionedRecord<T extends Record<string, unknown>>(value: T) {
  let current = { ...value };
  let version = 0;
  return {
    get version() { return version; },
    read(): { value: T; version: number } {
      return { value: { ...current }, version };
    },
    update(expectedVersion: number, next: T): void {
      if (expectedVersion !== version) throw new Error("stale record version");
      current = { ...next };
      version += 1;
    },
  };
}
