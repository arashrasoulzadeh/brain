import {
  createRouteDispatcher,
  requireProjectInput,
  requiredString,
  trimmedString,
  type PluginContext,
} from "../sdk-typescript/src/index.js";

type FunctionInput = {
  name?: unknown;
  input?: unknown;
  source?: unknown;
  version?: unknown;
  permission?: unknown;
};

export const dispatchRoute = createRouteDispatcher({ invoke, publish, call });

export function invoke(input: unknown): unknown {
  const payload = (input ?? {}) as FunctionInput;
  const name = requiredString(payload.name, "function name is required");
  return { name, input: payload.input ?? null, invoked: true };
}

export function publish(input: unknown): unknown {
  const payload = (input ?? {}) as FunctionInput;
  const name = trimmedString(payload.name);
  const source = typeof payload.source === "string" ? payload.source : "";
  const version = trimmedString(payload.version);
  if (!name || !source || !version)
    throw new Error("function name, source, and version are required");
  return { name, version, published: true };
}

export function call(input: unknown, context?: PluginContext): unknown {
  const payload = (context ? requireProjectInput(context, input) : input ?? {}) as FunctionInput;
  const name = requiredString(payload.name, "function name is required");
  const version = trimmedString(payload.version, "latest");
  const permission = trimmedString(payload.permission, "public");
  if (permission !== "public") throw new Error("function is not public");
  context?.plugins.observe.fire("callAccepted", {});
  return { name, version, input: payload.input ?? null, accepted: true, route: "public" };
}

