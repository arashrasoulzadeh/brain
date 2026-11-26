export type WorkerParams = {
  protocol: number;
  projectId: string;
  project?: ProjectContext;
  pluginId: string;
  version: string;
  operation: string;
  payload?: unknown;
};

export type RpcRequest = {
  jsonrpc: "2.0";
  id: string;
  method: "invoke";
  params: WorkerParams;
  /** Injected by the trusted worker boundary; absent on serialized fixture requests. */
  context?: PluginContext;
};

export type ProjectContext = Readonly<{ id: string; active?: boolean }>;

export type PluginContext = Readonly<{
  project: ProjectContext;
  plugin: Readonly<{ id: string; version: string }>;
  execution: Readonly<{ operation: string }>;
  plugins: Readonly<{ observe: Readonly<{ fire(event: string, payload: Readonly<Record<string, string | number | boolean>>): void }> }>;
}>;

export type Observation<T = unknown> = Readonly<{ projectId: string; pluginId: string; version: string; event: string; payload: T }>;

export type ObservationIntent = Readonly<{ event: string; payload: Readonly<Record<string, string | number | boolean>> }>;

export type ContextualPluginFunction = (context: PluginContext, input: unknown) => unknown;

export type RpcResponse = {
  jsonrpc: "2.0";
  id: string;
  result?: { protocol: number; success: boolean; payload?: unknown; observations?: readonly ObservationIntent[] };
  errorCode?: string;
};

export type PluginHandler = (request: RpcRequest, context?: PluginContext) => Promise<unknown> | unknown;

export type PluginRoute = {
  operation: string;
  function: string;
  capability: string;
  helper?: "sdk.dispatch" | "sdk.request" | "sdk.response" | "sdk.context" | "host.admin";
  response?: DeclarativeResponse;
};

export type DeclarativeResponse = {
  input: readonly {
    name: string;
    type: "string" | "value";
    trim?: boolean;
    required?: boolean;
    default?: string;
    error?: string;
  }[];
  output: Readonly<Record<string, string | number | boolean | null>>;
};

export type HostInvoker = {
  invoke(projectId: string, operation: string, capability: string, payload: unknown): unknown;
};

export type HostExecutionContext = {
  readonly projectId: string;
  readonly invocationId: string;
};

export type PluginFunction = (request: RpcRequest, context?: PluginContext) => unknown;
export { createRouteHandler, createHostHandler, createFunctionHandler, createOperationRoutes, createRequestDispatcher, createRouteDispatcher } from "./routes.js";
export { request, pluginRequest, createResourceRequest } from "./request.js";
export { validateProjectId, requireProjectInput, validateIdentifier, requiredString, trimmedString, validateBoundedInteger, requireAllowedValues } from "./validation.js";
export type { PayloadHandler } from "./routes.js";
export { createProjectMap } from "./project-store.js";
export { commitWithGrant, createCommitRollback, createTransactionGate, runCappedRetries, runWithTimeout } from "./execution.js";

export { createFixedWindowLimiter } from "./execution.js";
