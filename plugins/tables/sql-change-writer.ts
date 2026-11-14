import { validateIdentifier } from "../sdk-typescript/src/index.js";
import type { SQLChangeState } from "./sql-change-state.js";
export class SQLChangeWriter {
  constructor(private readonly state: SQLChangeState) {}
  begin(projectId: string, table: string, statement: string) {
    validateIdentifier(table);
    if (!projectId || !statement) throw new Error("invalid SQL change");
    let finished = false;
    const finish = (commit: boolean) => {
      if (finished) throw new Error("transaction already finished");
      finished = true;
      if (commit) this.state.changes.push({ projectId, table, statement });
    };
    return { commit: () => finish(true), rollback: () => finish(false) };
  }
}
