

export function customViewOrigin(origin: string, projectId: string) {
  const parsed = new URL(origin);
  if (parsed.protocol !== "https:" || !parsed.host || !projectId.trim())
    throw new Error("invalid custom view origin");
  return { origin: parsed.origin, projectId, isolated: true } as const;
}

export function restrictedFrame(origin: string, projectId: string, viewPath: string) {
  const parsed = new URL(origin);
  if (parsed.protocol !== "https:" || !projectId.trim() || !viewPath.startsWith("/"))
    throw new Error("invalid restricted frame");
  return { src: `${parsed.origin}${viewPath}`, projectId, sandbox: "allow-scripts" } as const;
}

const SENSITIVE_KEY = /(?:token|secret|password|credential|api[_-]?key|confirmation|destructive)/i;

export function rejectSensitiveBridgeData(value: unknown): void {
  if (!value || typeof value !== "object") return;
  if (Array.isArray(value)) {
    for (const item of value) rejectSensitiveBridgeData(item);
    return;
  }
  for (const [key, nested] of Object.entries(value)) {
    if (SENSITIVE_KEY.test(key)) throw new Error("invalid bridge message");
    rejectSensitiveBridgeData(nested);
  }
}

export type BridgeExecutionContext = {
  readonly projectId: string;
  readonly invocationId: string;
};

export type BridgeMessage = {
  readonly protocol: 1;
  readonly projectId: string;
  readonly invocationId: string;
  readonly operation: string;
  readonly payload?: unknown;
};

export function validateBridgeMessage(
  message: unknown,
  context: BridgeExecutionContext,
  operations: readonly string[],
): BridgeMessage {
  if (!message || typeof message !== "object" || !context.projectId || !context.invocationId) {
    throw new Error("invalid bridge message");
  }
  const candidate = message as Record<string, unknown>;
  const keys = Object.keys(candidate);
  if (
    candidate.protocol !== 1 ||
    candidate.projectId !== context.projectId ||
    candidate.invocationId !== context.invocationId ||
    typeof candidate.operation !== "string" ||
    !candidate.operation ||
    !operations.includes(candidate.operation) ||
    keys.some(
      (key) => !["protocol", "projectId", "invocationId", "operation", "payload"].includes(key),
    )
  ) {
    throw new Error("invalid bridge message");
  }
  rejectSensitiveBridgeData(candidate.payload);
  return candidate as BridgeMessage;
}

export function validateCsrfToken(expected: string, provided: string): void {
  if (!expected || !provided || expected.length !== provided.length)
    throw new Error("csrf validation failed");
  let difference = 0;
  for (let index = 0; index < expected.length; index += 1)
    difference |= expected.charCodeAt(index) ^ provided.charCodeAt(index);
  if (difference !== 0) throw new Error("csrf validation failed");
}
