import { deepEqual, equal, throws } from "node:assert/strict";
import { mkdtemp, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { spawnSync } from "node:child_process";
import { dispatchRoute as dispatchFile } from "../../plugins/files/index.js";
import { dispatchRoute as dispatchRoom } from "../../plugins/rooms/index.js";
import { dispatchRoute as dispatchFunction } from "../../plugins/functions/index.js";
import { dispatchRoute as dispatchTicket } from "../../plugins/matchmaking/index.js";

deepEqual(dispatchFunction("invoke", { name: " hello " }), {
  name: "hello",
  input: null,
  invoked: true,
});
throws(() => dispatchFunction("call", { name: "private", permission: "private" }), /not public/);
throws(() => dispatchFunction("constructor", {}), /not exported/);
deepEqual(dispatchTicket("queue", { id: "ticket-sdk-test" }), {
  id: "ticket-sdk-test",
  status: "queued",
});
deepEqual(dispatchTicket("cancel", { id: "ticket-sdk-test" }), {
  id: "ticket-sdk-test",
  status: "cancelled",
});

deepEqual(dispatchFile("upload", { key: " file ", data: "content" }), {
  key: "file",
  uploaded: true,
  published: false,
});
deepEqual(dispatchFile("metadata", { key: "file" }), {
  key: "file",
  published: false,
  size: 7,
});
throws(() => dispatchFile("download", { key: " " }), /file key is required/);
deepEqual(dispatchRoom("create", { id: "room-sdk-test" }), { id: "room-sdk-test", members: [] });
deepEqual(dispatchRoom("join", { id: "room-sdk-test", member: " player " }), {
  id: "room-sdk-test",
  member: "player",
  reconnected: false,
  members: ["player"],
});
throws(() => dispatchRoom("join", { id: "room-sdk-test" }), /room id and member are required/);

const directory = await mkdtemp(join(tmpdir(), "brain-manifest-helper-"));
try {
  const entrypoint = join(directory, "plugin.mjs");
  const descriptor = join(directory, "descriptor.json");
  await writeFile(entrypoint, 'export function dispatchRoute(name, input) { return { name, input }; }\nexport function handle() { throw new Error("legacy handler called"); }\n');
  await writeFile(descriptor, JSON.stringify({
    entrypoint,
    pluginId: "sample",
    version: "1",
    projectId: "project",
    routes: [{ operation: "sample.read", function: "read", helper: "sdk.dispatch" }],
  }));
  const params = { pluginId: "sample", version: "1", projectId: "project", operation: "sample.read", payload: { ok: true } };
  const lines = [
    { id: "ok", params },
    { id: "cross-project", params: { ...params, projectId: "other" } },
    { id: "undeclared", params: { ...params, operation: "sample.write" } },
  ].map((request) => JSON.stringify(request)).join("\n") + "\n";
  const worker = spawnSync(process.execPath, ["scripts/typescript-worker.mjs", descriptor], { input: lines, encoding: "utf8", timeout: 5000 });
  equal(worker.status, 0, worker.stderr);
  const [accepted, denied, undeclared] = worker.stdout.trim().split("\n").map((line) => JSON.parse(line) as { result?: unknown; errorCode?: string });
  deepEqual(accepted.result, { name: "read", input: { ok: true } });
  equal(denied.errorCode, "execution context mismatch");
  equal(undeclared.errorCode, "operation not declared in plugin manifest");
} finally {
  await rm(directory, { recursive: true, force: true });
}
