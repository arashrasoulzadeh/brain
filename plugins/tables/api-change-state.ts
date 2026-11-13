import type { ApiChange } from "./index.js";
export type ApiChangeState = { changes: ApiChange[] };
export function createApiChangeState(): ApiChangeState { return { changes: [] }; }
