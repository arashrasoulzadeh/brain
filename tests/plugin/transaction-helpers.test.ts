import { deepEqual, throws } from "node:assert/strict";
import { createCommitRollback } from "../../plugins/sdk-typescript/src/index.js";
import { createInvocationEffects } from "../../plugins/functions/index.js";
import { createApiChangeCapture, createDirectSQLChangeCapture, createTransactionalChangeCapture } from "../../plugins/tables/index.js";

const committed: string[] = [];
const tx = createCommitRollback(() => { committed.push("committed"); });
tx.commit();
throws(() => tx.commit(), /already finished/);
throws(() => tx.rollback(), /already finished/);
deepEqual(committed, ["committed"]);

const rolledBack = createCommitRollback(() => { committed.push("unexpected"); });
rolledBack.rollback();
throws(() => rolledBack.commit(), /already finished/);
deepEqual(committed, ["committed"]);

const effects = createInvocationEffects();
const invocation = effects.begin();
invocation.write("effect");
invocation.commit(false);
throws(() => invocation.write("late"), /invocation already finished/);
throws(() => invocation.commit(true), /invocation already finished/);
deepEqual(effects.list(), []);

const api = createApiChangeCapture();
const apiTx = api.begin("dev", "create", {});
apiTx.commit();
throws(() => apiTx.commit(), /already finished/);
deepEqual(api.list("prod"), []);
deepEqual(api.list("dev"), [{ projectId: "dev", operation: "create", payload: {} }]);

const sql = createDirectSQLChangeCapture();
sql.begin("dev", "records", "UPDATE records SET value = 1").rollback();
deepEqual(sql.list("dev"), []);

const rows = createTransactionalChangeCapture<number>();
const scoped = rows.beginScoped("dev", "records", 0, 1, ["owner"], ["member"]);
scoped.commit();
throws(() => scoped.rollback(), /already finished/);
deepEqual(rows.listScoped("prod"), []);
deepEqual(rows.listScoped("dev"), [{
  projectId: "dev",
  table: "records",
  oldValue: 0,
  newValue: 1,
  oldScope: ["owner"],
  newScope: ["member"],
}]);
