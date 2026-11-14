export type InvocationEffectsState = { committed: unknown[] };
export function createInvocationEffectsState(): InvocationEffectsState {
  return { committed: [] };
}
