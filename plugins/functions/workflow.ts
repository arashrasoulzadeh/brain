import { validateProjectId, runCappedRetries, requireProjectInput, type PluginContext } from "../sdk-typescript/src/index.js";
import { type createInvocationCommitStore } from "./invocation.js";

export type ProjectWorkflowAction = {
  readonly projectId: string;
  readonly id: string;
  readonly operation: string;
  readonly capability: string;
};

export function registerProjectWorkflowAction(
  actions: Map<string, ProjectWorkflowAction>,
  action: ProjectWorkflowAction,
): ProjectWorkflowAction {
  try {
    validateProjectId(action.projectId);
    if (!/^[a-zA-Z0-9][a-zA-Z0-9_-]*$/.test(action.id)) throw new Error("invalid action id");
  } catch {
    throw new Error("invalid project workflow action");
  }
  if (
    !action.operation ||
    !action.capability ||
    action.operation.startsWith("deploy.") ||
    action.operation.startsWith("release.") ||
    actions.has(`${action.projectId}:${action.id}`)
  ) {
    throw new Error("invalid project workflow action");
  }
  const registered = { ...action };
  actions.set(`${action.projectId}:${action.id}`, registered);
  return registered;
}

export function authorizeProjectWorkflowAction(
  actions: ReadonlyMap<string, ProjectWorkflowAction>,
  grants: ReadonlySet<string>,
  projectId: string,
  actionId: string,
): ProjectWorkflowAction {
  validateProjectId(projectId);
  const action = actions.get(`${projectId}:${actionId}`);
  if (!action || !grants.has(`${projectId}:${actionId}:${action.capability}`)) {
    throw new Error("workflow grant revoked");
  }
  return { ...action };
}

export type WorkflowCondition = { field: string; operator: string; value: unknown };
export function evaluateCondition(
  condition: WorkflowCondition,
  context: Readonly<Record<string, unknown>>,
): boolean {
  if (!condition.field || condition.operator !== "equals") throw new Error("unsupported condition");
  return (
    Object.prototype.hasOwnProperty.call(context, condition.field) &&
    Object.is(context[condition.field], condition.value)
  );
}

export function replaceProjectWorkflowStep(
  actions: Map<string, ProjectWorkflowAction>,
  projectId: string,
  stepId: string,
  replacement: ProjectWorkflowAction,
): ProjectWorkflowAction {
  const key = `${projectId}:${stepId}`;
  if (
    !/^[a-zA-Z0-9][a-zA-Z0-9_-]*$/.test(projectId) ||
    replacement.projectId !== projectId ||
    replacement.id !== stepId ||
    !actions.has(key) ||
    !replacement.operation ||
    replacement.operation.startsWith("deploy.") ||
    !replacement.capability
  ) {
    throw new Error("invalid workflow step replacement");
  }
  const updated = { ...replacement };
  actions.set(key, updated);
  return updated;
}

const inFlightWorkflows = new WeakMap<object, Map<string, Promise<unknown>>>();

export async function runIdempotentWorkflow<T>(
  store: ReturnType<typeof createInvocationCommitStore>,
  idempotencyKey: string,
  task: () => T | Promise<T>,
  maxAttempts: number,
): Promise<T> {
  if (store.has(idempotencyKey)) return store.result(idempotencyKey) as T;
  let storeWorkflows = inFlightWorkflows.get(store);
  if (!storeWorkflows) {
    storeWorkflows = new Map();
    inFlightWorkflows.set(store, storeWorkflows);
  }
  const existing = storeWorkflows.get(idempotencyKey);
  if (existing) return (await existing) as T;
  const execution = (async () => {
    const result = await runCappedRetries(task, maxAttempts);
    return store.commit(idempotencyKey, [], result) as T;
  })();
  storeWorkflows.set(idempotencyKey, execution);
  try {
    return await execution;
  } finally {
    storeWorkflows.delete(idempotencyKey);
  }
}

export function redactRunStatus<T>(status: T): T {
  const secret = /TOKEN|SECRET|PASSWORD|CREDENTIAL|API[_-]?KEY/i;
  const visit = (value: unknown): unknown => {
    if (Array.isArray(value)) return value.map(visit);
    if (!value || typeof value !== "object") return value;
    const result: Record<string, unknown> = {};
    for (const [key, nested] of Object.entries(value))
      result[key] = secret.test(key) ? "[redacted]" : visit(nested);
    return result;
  };
  return visit(status) as T;
}

export type WorkflowNode = { id: string; next: readonly string[] };
export function validateWorkflowGraph(nodes: readonly WorkflowNode[]): void {
  const byId = new Map(nodes.map((node) => [node.id, node]));
  if (
    nodes.length === 0 ||
    byId.size !== nodes.length ||
    nodes.some(
      (node) => !node.id.trim() || node.next.some((next) => !next.trim() || !byId.has(next)),
    )
  )
    throw new Error("invalid workflow graph");
  const visiting = new Set<string>();
  const visited = new Set<string>();
  const visit = (id: string): void => {
    if (visiting.has(id)) throw new Error("workflow cycle");
    if (visited.has(id)) return;
    visiting.add(id);
    for (const next of byId.get(id)!.next) visit(next);
    visiting.delete(id);
    visited.add(id);
  };
  for (const node of nodes) visit(node.id);
}

/** Project namespace is mandatory when an execution runner uses a shared commit store. */
export function runProjectWorkflow<T>(
  context: PluginContext,
  store: ReturnType<typeof createInvocationCommitStore>,
  idempotencyKey: string,
  task: (context: PluginContext) => T | Promise<T>,
  maxAttempts: number,
): Promise<T> {
  requireProjectInput(context, {});
  if (!idempotencyKey) throw new Error("workflow idempotency key required");
  return runIdempotentWorkflow(store, JSON.stringify([context.project.id, idempotencyKey]), () => task(context), maxAttempts);
}
