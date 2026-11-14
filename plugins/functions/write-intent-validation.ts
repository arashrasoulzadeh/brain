export type FunctionWriteIntent = { operation: "set" | "delete"; path: string; value?: unknown };
export function validateWriteIntents(
  intents: readonly FunctionWriteIntent[],
  declaredPaths: readonly string[],
): FunctionWriteIntent[] {
  const declared = new Set(declaredPaths);
  for (const intent of intents) {
    if (
      !intent ||
      !["set", "delete"].includes(intent.operation) ||
      !intent.path ||
      intent.path.includes("..") ||
      intent.path.startsWith("/") ||
      intent.path.endsWith("/") ||
      intent.path.includes("//")
    )
      throw new Error("invalid write intent");
    if (!declared.has(intent.path)) throw new Error("write intent not declared");
  }
  return intents.map((intent) => ({ ...intent }));
}
