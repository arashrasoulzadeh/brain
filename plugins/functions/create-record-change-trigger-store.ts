import { validateProjectId } from "../sdk-typescript/src/index.js";
import type { RecordChangeTrigger } from "./index.js";

export function createRecordChangeTriggerStore() {
  const triggers: RecordChangeTrigger[] = [];

  function match(projectId: string, collection: string, operation: string): RecordChangeTrigger[] {
    validateProjectId(projectId);
    return triggers
      .filter(
        (trigger) =>
          trigger.projectId === projectId &&
          trigger.collection === collection &&
          trigger.operation === operation,
      )
      .map((trigger) => ({ ...trigger, target: { ...trigger.target } }));
  }

  return {
    add(
      projectId: string,
      collection: string,
      operation: string,
      target: { functionId: string; version: number },
    ): RecordChangeTrigger {
      validateProjectId(projectId);
      if (!collection || !operation || !target.functionId || !Number.isSafeInteger(target.version) || target.version < 1)
        throw new Error("invalid record trigger");
      const trigger = { projectId, collection, operation, target: { ...target } };
      triggers.push(trigger);
      return { ...trigger, target: { ...trigger.target } };
    },
    match,
    dispatch(
      projectId: string,
      collection: string,
      operation: string,
      committed: boolean,
      ancestry: readonly string[] = [],
    ): RecordChangeTrigger[] {
      return committed
        ? match(projectId, collection, operation).filter((trigger) => !ancestry.includes(trigger.target.functionId))
        : [];
    },
  };
}
