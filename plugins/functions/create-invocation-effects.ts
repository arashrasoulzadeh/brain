import { createInvocationEffectsState } from "./invocation-effects-state.js";
import { InvocationEffectsWriter } from "./invocation-effects-writer.js";
import { InvocationEffectsReader } from "./invocation-effects-reader.js";
export function createInvocationEffects() {
  const state = createInvocationEffectsState();
  const writer = new InvocationEffectsWriter(state);
  const reader = new InvocationEffectsReader(state);
  return { begin: writer.begin.bind(writer), list: reader.list.bind(reader) };
}
