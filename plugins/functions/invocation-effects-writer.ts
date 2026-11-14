import type { InvocationEffectsState } from "./invocation-effects-state.js";
export class InvocationEffectsWriter {
  constructor(private readonly state: InvocationEffectsState) {}
  begin() {
    const staged: unknown[] = [];
    let finished = false;
    return {
      write: (effect: unknown) => {
        if (finished) throw new Error("invocation already finished");
        staged.push(effect);
      },
      commit: (success: boolean) => {
        if (finished) throw new Error("invocation already finished");
        finished = true;
        if (success) this.state.committed.push(...staged);
      },
    };
  }
}
