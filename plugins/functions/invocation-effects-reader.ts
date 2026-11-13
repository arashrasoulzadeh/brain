import type { InvocationEffectsState } from "./invocation-effects-state.js";
export class InvocationEffectsReader { constructor(private readonly state: InvocationEffectsState) {} list(): unknown[] { return [...this.state.committed]; } }
