import { validateProjectId } from "../sdk-typescript/src/index.js";
import type { FunctionSummary } from "./types.js";

export function functionManagementPage(projectId: string, functions: readonly FunctionSummary[]) {
  validateProjectId(projectId, "invalid function management");
  if (functions.some((item) => !item.id || !Number.isSafeInteger(item.version) || !item.status))
    throw new Error("invalid function management");
  return { projectId, title: "Functions", functions: functions.map((item) => ({ ...item })) };
}
