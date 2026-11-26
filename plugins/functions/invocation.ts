import { createTransactionGate, validateProjectId, pluginRequest, type RpcRequest } from "../sdk-typescript/src/index.js";
import { isIP } from "node:net";

export function createInvocationCommitStore() {
  const records = new Map<string, unknown>();
  const committedEffects: unknown[] = [];
  let lifecycleRevision = 0;
  return {
    commit(
      idempotencyKey: string,
      effects: readonly unknown[],
      result: unknown,
      success = true,
      expectedRevision = lifecycleRevision,
    ): unknown {
      if (!idempotencyKey) throw new Error("idempotency key required");
      if (records.has(idempotencyKey)) return records.get(idempotencyKey);
      if (expectedRevision !== lifecycleRevision)
        throw new Error("stale invocation lifecycle revision");
      if (!success) throw new Error("invocation commit failed");
      committedEffects.push(...effects);
      records.set(idempotencyKey, result);
      return result;
    },
    effects(): unknown[] {
      return [...committedEffects];
    },
    result(idempotencyKey: string): unknown {
      return records.get(idempotencyKey);
    },
    has(idempotencyKey: string): boolean {
      return records.has(idempotencyKey);
    },
    setLifecycleRevision(revision: number): void {
      lifecycleRevision = revision;
    },
  };
}

export function createInvocationEffects() {
  const committed: unknown[] = [];
  return {
    begin() {
      const staged: unknown[] = [];
      const gate = createTransactionGate("invocation already finished");
      return {
        write(effect: unknown): void {
          gate.ensureOpen();
          staged.push(effect);
        },
        commit(success: boolean): void {
          gate.finish(() => {
            if (success) committed.push(...staged);
          });
        },
      };
    },
    list(): unknown[] {
      return [...committed];
    },
  };
}

export type FunctionInvocationRequest = {
  projectId: string;
  functionId: string;
  version: string;
  inputs: Readonly<Record<string, unknown>>;
  idempotencyKey: string;
};
export type FunctionInvocationResult = {
  success: boolean;
  output?: unknown;
  error?: string;
  diagnostics: readonly string[];
};
export type HttpsFetcher = (input: string, init?: RequestInit) => Promise<Response>;

export function createInvocationRequest(
  input: FunctionInvocationRequest,
): FunctionInvocationRequest {
  validateProjectId(input.projectId);
  if (
    !input.functionId ||
    !input.version ||
    !input.idempotencyKey ||
    !input.inputs ||
    typeof input.inputs !== "object"
  )
    throw new Error("invalid function invocation request");
  return { ...input, inputs: { ...input.inputs } };
}

export function loadDeclaredInputs(
  inputs: Readonly<Record<string, unknown>>,
  declarations: readonly string[],
): Readonly<Record<string, unknown>> {
  if (
    !inputs ||
    typeof inputs !== "object" ||
    declarations.some((name) => !/^[A-Za-z_][A-Za-z0-9_]*$/.test(name))
  )
    throw new Error("invalid input declaration");
  const loaded: Record<string, unknown> = {};
  for (const name of declarations) {
    if (!Object.prototype.hasOwnProperty.call(inputs, name))
      throw new Error("declared input missing");
    loaded[name] = inputs[name];
  }
  return loaded;
}

export function functionRequest(
  projectId: string,
  pluginId: string,
  version: string,
  operation: "functions.invoke" | "functions.publish",
  payload: unknown = {},
): RpcRequest {
  return pluginRequest("functions", projectId, pluginId, version, operation, payload);
}

export function validateInvocationResult(
  result: FunctionInvocationResult,
  maxDiagnostics = 20,
): FunctionInvocationResult {
  if (
    typeof result.success !== "boolean" ||
    result.diagnostics.length > maxDiagnostics ||
    result.diagnostics.some((line) => line.length > 4096) ||
    (result.success && result.error) ||
    (!result.success && !result.error)
  )
    throw new Error("invalid function invocation result");
  return { ...result, diagnostics: [...result.diagnostics] };
}

export function parseInvocationResult(value: unknown): FunctionInvocationResult {
  if (!value || typeof value !== "object") throw new Error("malformed function output");
  const result = value as Partial<FunctionInvocationResult>;
  if (
    Object.keys(value).some((key) => !["success", "output", "error", "diagnostics"].includes(key))
  )
    throw new Error("extra function result rejected");
  if (!Array.isArray(result.diagnostics)) throw new Error("malformed function output");
  return validateInvocationResult(result as FunctionInvocationResult);
}

export async function invokeDirectHttps(
  endpoint: string,
  input: unknown,
  fetcher: HttpsFetcher = fetch,
): Promise<FunctionInvocationResult> {
  const target = new URL(endpoint);
  if (target.protocol !== "https:" || target.username || target.password)
    throw new Error("HTTPS endpoint required");
  const hostname = target.hostname.toLowerCase();
  const normalizedIPv6Host = hostname.replace(/^\[|\]$/g, "");
  const addressType = isIP(normalizedIPv6Host);
  const privateIPv4 =
    addressType === 4 &&
    (/^(10|127)\./.test(normalizedIPv6Host) ||
      /^192\.168\./.test(normalizedIPv6Host) ||
      /^172\.(1[6-9]|2[0-9]|3[0-1])\./.test(normalizedIPv6Host) ||
      normalizedIPv6Host === "169.254.169.254");
  const privateIPv6 =
    addressType === 6 &&
    (normalizedIPv6Host === "::1" ||
      normalizedIPv6Host.startsWith("::ffff:") ||
      normalizedIPv6Host.startsWith("fc") ||
      normalizedIPv6Host.startsWith("fd") ||
      normalizedIPv6Host.startsWith("fe8") ||
      normalizedIPv6Host.startsWith("fe9") ||
      normalizedIPv6Host.startsWith("fea") ||
      normalizedIPv6Host.startsWith("feb"));
  const ambiguousNumericHost = addressType === 0 && /^[0-9.]+$/.test(hostname);
  if (!hostname || hostname === "localhost" || privateIPv4 || privateIPv6 || ambiguousNumericHost)
    throw new Error("HTTPS endpoint target denied");
  const response = await fetcher(target.toString(), {
    method: "POST",
    redirect: "error",
    headers: { "content-type": "application/json" },
    body: JSON.stringify(input),
  });
  if (!response.ok) throw new Error(`HTTPS invocation failed: ${response.status}`);
  return parseInvocationResult(await response.json());
}

export async function retryInvocation<T>(
  store: ReturnType<typeof createInvocationCommitStore>,
  idempotencyKey: string,
  execute: () => T | Promise<T>,
  effects: readonly unknown[] = [],
): Promise<T> {
  if (store.has(idempotencyKey)) return store.result(idempotencyKey) as T;
  const result = await execute();
  return store.commit(idempotencyKey, effects, result) as T;
}

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

export function executeFixture(fixture: () => unknown): FunctionInvocationResult {
  try {
    return parseInvocationResult(fixture());
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : "fixture execution failed",
      diagnostics: [],
    };
  }
}
