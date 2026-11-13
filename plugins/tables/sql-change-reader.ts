import type { SQLChangeState } from "./sql-change-state.js";
export class SQLChangeReader { constructor(private readonly state: SQLChangeState) {} list(projectId: string) { return this.state.changes.filter((change) => change.projectId === projectId).map((change) => ({ ...change })); } }
