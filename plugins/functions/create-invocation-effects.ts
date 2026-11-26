export function createInvocationEffects() {
  const committed: unknown[] = [];
  return {
    begin() {
      const staged: unknown[] = [];
      let finished = false;
      return {
        write(effect: unknown): void {
          if (finished) throw new Error("invocation already finished");
          staged.push(effect);
        },
        commit(success: boolean): void {
          if (finished) throw new Error("invocation already finished");
          finished = true;
          if (success) committed.push(...staged);
        },
      };
    },
    list(): unknown[] {
      return [...committed];
    },
  };
}
