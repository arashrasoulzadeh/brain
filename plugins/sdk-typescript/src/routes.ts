export type PayloadHandler = (input: unknown) => unknown;

export function createRouteDispatcher(handlers: Readonly<Record<string, PayloadHandler>>) {
  const registered = Object.freeze({ ...handlers });
  return (name: string, input: unknown): unknown => {
    if (!Object.hasOwn(registered, name) || typeof registered[name] !== "function")
      throw new Error("declared plugin function is not exported: " + name);
    return registered[name](input);
  };
}

export function trimmedString(value: unknown, fallback = ""): string {
  return typeof value === "string" ? value.trim() : fallback;
}

export function requiredString(value: unknown, message: string): string {
  const result = trimmedString(value);
  if (!result) throw new Error(message);
  return result;
}

export function createOperationRoutes<Operation extends string>(
  operations: readonly Operation[],
  execute: (operation: Operation, input: unknown) => unknown,
) {
  const handlers = Object.freeze(
    Object.fromEntries(
      operations.map((operation) => [operation, (input: unknown) => execute(operation, input)]),
    ) as Record<Operation, PayloadHandler>,
  );
  return { handlers, dispatch: createRouteDispatcher(handlers) };
}
