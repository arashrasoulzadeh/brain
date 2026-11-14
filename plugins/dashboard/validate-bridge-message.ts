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
import { rejectSensitiveBridgeData } from "./reject-sensitive-bridge-data.js";
