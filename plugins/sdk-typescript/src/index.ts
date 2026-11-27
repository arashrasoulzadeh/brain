import type { PureHelperName } from "./helpers.mjs";

export type WorkerParams = {
  protocol: number;
  projectId: string;
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

export type ProjectContext = Readonly<{ id: string; active?: true }>;

export type PluginContext = Readonly<{
  project: ProjectContext;
  plugin: Readonly<{ id: string; version: string }>;
  execution: Readonly<{ operation: string; invocationId?: string }>;
  config?: Readonly<Record<string, string | number | boolean>>;
  helpers?: Readonly<{ call: (name: PureHelperName, input: unknown) => unknown }>;
  plugins: Readonly<{ has?: (id: string) => boolean; observe: Readonly<{ fire(event: string, payload: Readonly<Record<string, string | number | boolean>>): void }> }>;
}>;

export type Observation<T = unknown> = Readonly<{ projectId: string; pluginId: string; version: string; event: string; payload: T }>;

export type ContextualPluginFunction = (context: PluginContext, input: unknown) => unknown;

export type RpcResponse = {
  jsonrpc: "2.0";
  id: string;
  result?: { protocol: number; success: boolean; payload?: unknown };
  errorCode?: string;
};

export type PluginHandler = (request: RpcRequest, context?: PluginContext) => Promise<unknown> | unknown;

export type PluginRoute = {
  operation: string;
  function: string;
  capability: string;
  resourceField?: string;
  helper?: "sdk.dispatch" | "sdk.request" | "sdk.response" | "sdk.context" | "sdk.pipeline" | "host.model" | "host.call";
  call?: Readonly<{ plugin: string; version: string; operation: string }>;
  response?: DeclarativeResponse;
  pipeline?: DeclarativePipeline;
  model?: Readonly<{ name: string; action: "create" | "read" | "list" | "update" | "delete" | "transition"; transition?: string }>;
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
export { validateProjectId, validateIdentifier, requiredString, trimmedString, validateBoundedInteger, requireAllowedValues } from "./validation.js";
export type { PayloadHandler } from "./routes.js";
export { createProjectMap } from "./project-store.js";
export { commitWithGrant, createCommitRollback, createTransactionGate, runCappedRetries, runWithTimeout } from "./execution.js";

export { createFixedWindowLimiter } from "./execution.js";

export type DeclarativePipeline = Readonly<{
  steps: readonly Readonly<{ helper: PureHelperName | "observe.fire"; input: string; save: string }>[];
  return: string;
}>;
export { validateModel, transitionState } from "./declarative.js";
export type { ModelSchema, StateMachine } from "./declarative.js";

export { callHelper, hasHelper, helperNames } from "./helpers.mjs";
export type { PureHelperName } from "./helpers.mjs";
