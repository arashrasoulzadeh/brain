import { validateProjectId } from "../sdk-typescript/src/index.js";
import type { FunctionInvocationRequest } from "./invocation-types.js";
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
