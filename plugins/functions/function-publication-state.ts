import type { FunctionPublication } from "./index.js";
export type FunctionPublicationState = {
  publications: Map<string, FunctionPublication>;
  activeVersions: Map<string, number>;
};
export function createFunctionPublicationState(): FunctionPublicationState {
  return { publications: new Map(), activeVersions: new Map() };
}
