import { createTransactionGate } from "../sdk-typescript/src/index.js";

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
