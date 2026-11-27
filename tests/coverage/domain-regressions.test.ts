import assert from "node:assert/strict";
import { test } from "node:test";
import { QuotaReader, QuotaReserver, QuotaReleaser } from "../../plugins/files/quota.js";
import { FileGrantStore, FileAccessReader, FileAccessDeleter, configurePrivateBucket, protectedDownloadHeaders, requireExplicitPublication } from "../../plugins/files/access.js";
import { MatchResultReader, MatchResultSaver, validateMatchResult } from "../../plugins/arena/match-results.js";
import { MatchmakingConfigReader, MatchmakingConfigStatus, MatchmakingConfigWriter, validateConfig } from "../../plugins/matchmaking/config.js";
import { TableExposureReader, TableExposureWriter } from "../../plugins/tables/access.js";
import { createVersionedRecord, paginate, validateFilter, validateValue } from "../../plugins/tables/records.js";
import { createSchemaPublicationAudit, planSchemaChange, previewSchemaSQL, requireRecentSchemaAuth } from "../../plugins/tables/schema.js";
import { createManifestRequest, createProjectLeaseStore, createProjectSet, type PluginContext } from "../../plugins/sdk-typescript/src/index.js";
import * as functions from "../../plugins/functions/routes.js";

test("quota limits and access grants are independent for each project", () => {
  const reserved = new Map<string, number>();
  const reader = new QuotaReader(reserved);
  const reserve = new QuotaReserver(reserved, new Map([["alpha", 10], ["beta", 3]]));
  const release = new QuotaReleaser(reserved);
  assert.equal(reader.reservedFor("alpha"), 0);
  reserve.reserve("alpha", 7);
  reserve.reserve("alpha", 3);
  assert.equal(reader.reservedFor("alpha"), 10);
  assert.throws(() => reserve.reserve("alpha", 1), /quota exceeded/);
  assert.throws(() => reserve.reserve("unknown", 0), /quota exceeded/);
  for (const size of [-1, 0.5, Infinity]) {
    assert.throws(() => reserve.reserve("alpha", size), /invalid quota/);
    assert.throws(() => release.release("alpha", size), /invalid quota/);
  }
  release.release("beta", 2);
  release.release("alpha", 20);
  assert.equal(reader.reservedFor("beta"), 0);
  assert.equal(reader.reservedFor("alpha"), 0);
  const grants = new Set<string>();
  const grant = new FileGrantStore(grants);
  const read = new FileAccessReader(grants);
  const remove = new FileAccessDeleter(grants);
  assert.throws(() => grant.grant("alpha", "", "user"), /invalid file grant/);
  assert.throws(() => grant.grant("alpha", "file", ""), /invalid file grant/);
  grant.grant("alpha", "file", "user");
  assert.doesNotThrow(() => read.read("alpha", "file", "user"));
  assert.doesNotThrow(() => remove.delete("alpha", "file", "user"));
  for (const access of [read.read.bind(read), remove.delete.bind(remove)]) {
    assert.throws(() => access("alpha", "file", " "), /invalid file actor/);
    assert.throws(() => access("beta", "file", "user"), /access denied/);
  }
  const bucket = { endpoint: "https://storage.example", bucket: "private", region: "test" };
  assert.deepEqual(configurePrivateBucket(bucket), { ...bucket, publicAccess: false });
  for (const key of ["endpoint", "bucket", "region"] as const) {
    assert.throws(() => configurePrivateBucket({ ...bucket, [key]: "" }), /invalid bucket/);
  }
  assert.throws(() => configurePrivateBucket({ ...bucket, endpoint: "http://storage.example" }), /HTTPS/);
  for (const name of ["", "a\rb", "a\nb"]) assert.throws(() => protectedDownloadHeaders(name), /invalid download/);
  assert.equal(protectedDownloadHeaders('a"b')["content-disposition"], 'attachment; filename="ab"');
  assert.throws(() => requireExplicitPublication(false), /explicit publication/);
  requireExplicitPublication(true);
});

test("match results and matchmaking configurations enforce limits and return copies", () => {
  const result = { projectId: "alpha", roomId: "room", resultId: "alpha:result", completedAt: "2026-01-01T00:00:00Z", winnerIds: ["one"] };
  for (const invalid of [{ resultId: "invalid" }, { resultId: "beta:result" }, { completedAt: "" }, { completedAt: "invalid" }, { winnerIds: [] }]) {
    assert.throws(() => validateMatchResult({ ...result, ...invalid }));
  }
  const results = new Map<string, typeof result>();
  const reader = new MatchResultReader(results);
  const saver = new MatchResultSaver(results);
  assert.equal(reader.get("absent"), undefined);
  saver.save(result);
  result.winnerIds.push("two");
  assert.deepEqual(reader.get(result.resultId)?.winnerIds, ["one"]);
  assert.throws(() => saver.save(result), /already exists/);
  const config = { projectId: "alpha", scope: "ranked", matchSize: 2, ticketTtlSeconds: 30, enabled: true };
  for (const invalid of [{ projectId: "bad/id" }, { scope: "1scope" }, { matchSize: 0 }, { matchSize: 65 }, { ticketTtlSeconds: 0 }, { ticketTtlSeconds: 3601 }]) {
    assert.throws(() => validateConfig({ ...config, ...invalid }), /invalid matchmaking/);
  }
  const configs = new Map<string, typeof config>();
  const status = new MatchmakingConfigStatus(configs);
  const writer = new MatchmakingConfigWriter(configs);
  const configReader = new MatchmakingConfigReader(configs);
  assert.equal(status.isEnabled("alpha", "ranked"), false);
  assert.throws(() => configReader.get("alpha", "ranked"), /not found/);
  writer.set(config);
  const copy = configReader.get("alpha", "ranked");
  copy.enabled = false;
  assert.equal(status.isEnabled("alpha", "ranked"), true);
  assert.equal(status.isEnabled("beta", "ranked"), false);
});

test("table schema, exposure and pagination validate edge inputs and isolate records", () => {
  const exposure = new Map<string, { projectId: string; table: string; columns: readonly string[] }>();
  const reader = new TableExposureReader(exposure);
  const writer = new TableExposureWriter(exposure);
  assert.equal(reader.get("alpha", "items"), undefined);
  assert.throws(() => reader.get("alpha", "invalid/table"), /invalid table exposure/);
  assert.throws(() => writer.expose("alpha", "items", []), /invalid table exposure/);
  assert.throws(() => writer.expose("alpha", "invalid/table", ["id"]), /invalid table exposure/);
  writer.expose("alpha", "items", ["id"]);
  assert.deepEqual(reader.get("alpha", "items")?.columns, ["id"]);
  assert.equal(reader.get("beta", "items"), undefined);
  const record = createVersionedRecord({ count: 1 });
  const first = record.read();
  first.value.count = 99;
  assert.deepEqual(record.read(), { value: { count: 1 }, version: 0 });
  record.update(0, { count: 2 });
  assert.deepEqual(record.read(), { value: { count: 2 }, version: 1 });
  for (const limit of [0, 101, 1.5]) assert.throws(() => paginate([1], undefined, limit), /invalid page limit/);
  for (const cursor of ["-1", "3", "0.5", "invalid"]) assert.throws(() => paginate([1, 2], Buffer.from(cursor).toString("base64url"), 1), /invalid page cursor/);
  const page = paginate([1, 2], undefined, 1);
  assert.deepEqual(paginate([1, 2], page.nextCursor, 1), { items: [2] });
  for (const filter of [{ other: { eq: 1 } }, { id: 1 }, { id: null }, { id: { unsafe: 1 } }]) assert.throws(() => validateFilter(filter, ["id"]), /unapproved filter/);
  validateFilter({ id: { eq: 1 } }, ["id"]);
  for (const value of [undefined, () => 1, "x".repeat(4097)]) assert.throws(() => validateValue("id", value), /invalid value/);
  validateValue("id", "ok");
  assert.throws(() => planSchemaChange("alpha", "items", { kind: "create_index", name: "index", columns: [] }), /index columns required/);
  const index = planSchemaChange("alpha", "items", { kind: "create_index", name: "index", columns: ["id", "value"] });
  assert.equal(previewSchemaSQL(index), 'CREATE INDEX "index" ON "project_alpha"."items" ("id", "value");');
  for (const [at, now] of [[NaN, 0], [0, Infinity], [2, 1], [0, 300001]]) assert.throws(() => requireRecentSchemaAuth(at, now), /recent authentication/);
  requireRecentSchemaAuth(0, 300000);
  const audit = createSchemaPublicationAudit();
  for (const [project, actor, time] of [["beta", "operator", 1], ["alpha", "", 1], ["alpha", "operator", NaN]] as const) {
    assert.throws(() => audit.record(project, actor, index, time), /invalid schema publication/);
  }
  audit.record("alpha", "operator", index, 1);
  assert.equal(audit.list("beta").length, 0);
  assert.deepEqual(audit.list("alpha")[0].change, index.change);
});

test("manifest serializers and project lease helpers bind resource identity and reject invalid leases", () => {
  const request = createManifestRequest("rooms", "id");
  assert.equal(request("alpha", "1", "rooms.create", "room").params.pluginId, "rooms");
  assert.deepEqual(request("alpha", "1", "rooms.create", "room", { id: "forged" }).params.payload, { id: "room" });
  const add = createProjectSet<string>();
  assert.equal(add("alpha", "key", "value"), true);
  assert.equal(add("alpha", "key", "value"), false);
  assert.equal(add("beta", "key", "value"), true);
  const leases = createProjectLeaseStore();
  for (const [key, worker, now, duration] of [["", "worker", 0, 1], ["key", "", 0, 1], ["key", "worker", NaN, 1], ["key", "worker", 0, 0.5], ["key", "worker", 0, 0], ["key", "worker", Number.MAX_SAFE_INTEGER, 1]] as const) {
    assert.throws(() => leases.acquire("alpha", key, worker, now, duration), /invalid job lease/);
  }
  assert.equal(leases.renew("alpha", "key", "worker", 0, 10), false);
  assert.equal(leases.acquire("alpha", "key", "worker", 0, 10), true);
  assert.equal(leases.acquire("alpha", "key", "other", 1, 10), false);
  assert.equal(leases.renew("alpha", "key", "other", 1, 10), false);
  assert.equal(leases.renew("alpha", "key", "worker", 10, 10), false);
  assert.equal(leases.acquire("alpha", "key", "other", 10, 10), true);
  assert.equal(leases.renew("alpha", "key", "other", 11, 10), true);
  assert.equal(leases.acquire("beta", "key", "worker", 11, 10), true);
});

test("function dispatch rejects missing names and validates public calls and publication", () => {
  const events: string[] = [];
  const ctx: PluginContext = { project: { id: "alpha", active: true }, plugin: { id: "functions", version: "1" }, execution: { operation: "functions.call" }, plugins: { observe: { fire(event) { events.push(event); } } } };
  assert.deepEqual(functions.invoke({ name: "echo" }), { name: "echo", input: null, invoked: true });
  assert.throws(() => functions.invoke(null), /name is required/);
  for (const input of [{}, { name: "echo", source: 1, version: "1" }, { name: "echo", source: "source" }]) assert.throws(() => functions.publish(input), /name, source, and version/);
  assert.deepEqual(functions.publish({ name: " echo ", source: "source", version: " 1 " }), { name: "echo", version: "1", published: true });
  assert.throws(() => functions.call({ projectId: "beta", name: "echo" }, ctx), /project context mismatch/);
  assert.throws(() => functions.call({ name: "echo", permission: "private" }, ctx), /not public/);
  assert.deepEqual(functions.call({ name: "echo", input: false }, ctx), { name: "echo", version: "latest", input: false, accepted: true, route: "public" });
  assert.deepEqual(events, ["callAccepted"]);
});
