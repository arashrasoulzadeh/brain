import type { FunctionInvocationResult } from "./invocation-types.js";
import { parseInvocationResult } from "./parse-invocation-result.js";
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
