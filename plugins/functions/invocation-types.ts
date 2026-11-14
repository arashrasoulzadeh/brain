export type FunctionInvocationRequest = { projectId: string; functionId: string; version: string; inputs: Readonly<Record<string, unknown>>; idempotencyKey: string };
export type FunctionInvocationResult = { success: boolean; output?: unknown; error?: string; diagnostics: readonly string[] };
export type HttpsFetcher = (input: string, init?: RequestInit) => Promise<Response>;
