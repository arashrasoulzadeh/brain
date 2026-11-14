import { validateProjectId } from "../sdk-typescript/src/index.js";
export class FunctionVersionEnqueuer {
  enqueueVersion(
    projectId: string,
    functionId: string,
    functionVersion: number,
    enqueuedAt: number,
  ) {
    validateProjectId(projectId);
    if (
      !functionId ||
      !Number.isSafeInteger(functionVersion) ||
      functionVersion < 1 ||
      !Number.isSafeInteger(enqueuedAt) ||
      enqueuedAt < 0
    )
      throw new Error("invalid function enqueue");
    return { projectId, functionId, functionVersion, enqueuedAt };
  }
}
