import type { FunctionDraft } from "./index.js";
export type FunctionDraftState = { drafts: Map<string, FunctionDraft> };
export function createFunctionDraftState(): FunctionDraftState { return { drafts: new Map() }; }
