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
