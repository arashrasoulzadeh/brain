import type { DirectSQLChange } from "./index.js";
export type SQLChangeState = { changes: DirectSQLChange[] };
export function createSQLChangeState(): SQLChangeState { return { changes: [] }; }
