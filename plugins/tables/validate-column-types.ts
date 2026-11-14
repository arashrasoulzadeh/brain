export function validateColumnTypes(columns: Record<string, string>): void {
  const supported = new Set(["string", "number", "boolean", "date"]);
  if (Object.values(columns).some((type) => !supported.has(type)))
    throw new Error("unsupported column type");
}
