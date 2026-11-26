import type { PluginFunction, RpcRequest } from "./index.js";

export type PayloadHandler = (input: unknown) => unknown;

export function createRouteDispatcher(handlers: Readonly<Record<string, PayloadHandler>>) {
  return createDispatcher<unknown>(handlers);
}

export function createRequestDispatcher(
  handlers: Readonly<Record<string, PluginFunction>>,
  missingHandlerError?: string,
) {
  return createDispatcher<RpcRequest>(handlers, missingHandlerError);
}

function createDispatcher<Input>(
  handlers: Readonly<Record<string, (input: Input) => unknown>>,
  missingHandlerError?: string,
) {
  const registered = Object.freeze({ ...handlers });
  return (name: string, input: Input): unknown => {
    if (!Object.hasOwn(registered, name) || typeof registered[name] !== "function")
      throw new Error(missingHandlerError ?? "declared plugin function is not exported: " + name);
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
