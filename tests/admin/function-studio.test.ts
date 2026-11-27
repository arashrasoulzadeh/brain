import assert from "node:assert/strict";
import { test } from "node:test";
import {
  byteLength,
  catalogStats,
  checkedFunctionDraft,
  checkedPublication,
  checkedTestResult,
  formatFunctionDate,
  functionFormError,
  functionTemplates,
  loadFunctionDrafts,
  MAX_INPUT_BYTES,
  MAX_SOURCE_BYTES,
  normalizePermissions,
  parseTestInput,
  publicationURL,
  resolveFunctionScope,
  sendFunctionCommand,
  type FunctionScope,
} from "../../admin-nextjs/app/lib/function-studio.js";

const scope: FunctionScope = { project: "immutable", namespace: "alias", token: "fixture" };
const draft = { project_id: "immutable", name: "echo", version: 1, source: functionTemplates.echo.source, created_at: "2026-01-01" };

test("function studio validates form boundaries and starter modules", () => {
  for (const template of Object.values(functionTemplates)) {
    assert.equal(functionFormError("echo", template.source, "1"), "");
    assert.doesNotThrow(() => parseTestInput(template.input));
  }
  for (const name of ["", "1hello", " hello", "a/b", "a".repeat(65)]) {
    assert.match(functionFormError(name, draft.source, "1"), /Name must/);
  }
  assert.match(functionFormError("echo", " ", "1"), /source is required/);
  assert.match(functionFormError("echo", "é".repeat(MAX_SOURCE_BYTES / 2 + 1), "1"), /exceeds/);
  assert.equal(byteLength("é"), 2);
  for (const version of ["0", "-1", "1.5", "1e2", "9007199254740992"]) {
    assert.match(functionFormError("echo", draft.source, version), /positive safe integer/);
  }
  assert.deepEqual(normalizePermissions(" b, a, b, ,"), ["a", "b"]);
  assert.deepEqual(normalizePermissions(""), []);
  assert.equal(parseTestInput("null"), null);
  assert.deepEqual(parseTestInput("[1,2]"), [1, 2]);
  assert.throws(() => parseTestInput("{"), /valid JSON/);
  assert.throws(() => parseTestInput(" ".repeat(MAX_INPUT_BYTES + 1)), /exceeds/);
});

test("studio response validation rejects incorrect project, function and false success", () => {
  assert.deepEqual(checkedFunctionDraft(draft, scope, "echo"), draft);
  assert.throws(() => checkedFunctionDraft(null, scope, "echo"), /does not match/);
  assert.throws(() => checkedFunctionDraft({ ...draft, project_id: "foreign" }, scope, "echo"), /does not match/);
  assert.throws(() => checkedFunctionDraft({ ...draft, name: "other" }, scope, "echo"), /does not match/);
  assert.throws(() => checkedFunctionDraft({ ...draft, source: 1 }, scope, "echo"), /does not match/);
  assert.throws(() => checkedFunctionDraft({ ...draft, version: "1" }, scope, "echo"), /does not match/);
  const validated = { valid: true, project_id: "immutable", name: "echo" };
  assert.equal(checkedTestResult(validated, scope, "echo", false), true);
  assert.equal(checkedTestResult({ ...validated, result: null }, scope, "echo", true), null);
  assert.deepEqual(checkedTestResult({ ...validated, result: { ok: true } }, scope, "echo", true), { ok: true });
  for (const result of [null, { ...validated, valid: false }, { ...validated, project_id: "other" }, { ...validated, name: "other" }, validated]) {
    assert.throws(() => checkedTestResult(result, scope, "echo", true), /did not confirm/);
  }
  const publication = { published: true, project_id: "immutable", name: "echo", version: 2, method: "POST", published_at: "2026-01-01" };
  assert.deepEqual(checkedPublication(publication, scope, "echo"), { version: 2, method: "POST", published_at: "2026-01-01" });
  for (const result of [null, { ...publication, published: false }, { ...publication, project_id: "other" }, { ...publication, name: "other" }, { ...publication, version: null }, { ...publication, method: "" }, { ...publication, published_at: "" }]) {
    assert.throws(() => checkedPublication(result, scope, "echo"), /not confirmed/);
  }
});

test("catalog shows actual current publication and tolerates missing timestamps", () => {
  assert.deepEqual(catalogStats([draft, { ...draft, public: true, published_at: "2026-01-02" }]), { total: 2, published: 1, unpublished: 1, public: 1 });
  assert.equal(publicationURL("alias", "a b"), "/alias/functions/a%20b");
  assert.equal(formatFunctionDate(), "—");
  assert.equal(formatFunctionDate("invalid"), "—");
  assert.notEqual(formatFunctionDate("2026-01-01"), "—");
});

test("studio requests bind the canonical project and preserve host rejection", async () => {
  const original = globalThis.fetch;
  const controller = new AbortController();
  const calls: { url: string; init?: RequestInit }[] = [];
  let response = Response.json([{ ID: "immutable", Slug: "alias", Active: true }]);
  globalThis.fetch = async (input, init) => {
    calls.push({ url: String(input), init });
    return response;
  };
  try {
    assert.deepEqual(await resolveFunctionScope("fixture", "alias", controller.signal), scope);
    assert.equal(calls[0].url, "/_system/admin/projects");
    assert.equal(calls[0].init?.signal, controller.signal);
    await assert.rejects(resolveFunctionScope("", "alias", controller.signal), /sign in/);
    await assert.rejects(resolveFunctionScope("fixture", undefined, controller.signal), /sign in/);
    response = Response.json([{ ID: "immutable", Active: false }]);
    await assert.rejects(resolveFunctionScope("fixture", "immutable", controller.signal), /Project unavailable/);
    response = Response.json([{ ID: "immutable", Active: true }]);
    assert.deepEqual(await resolveFunctionScope("fixture", "immutable", controller.signal), { ...scope, namespace: "immutable" });
    response = new Response("", { status: 403 });
    await assert.rejects(resolveFunctionScope("fixture", "alias", controller.signal), /Projects unavailable/);
    response = Response.json([draft, { ...draft, project_id: "foreign" }]);
    assert.deepEqual(await loadFunctionDrafts(scope, controller.signal), [draft]);
    response = Response.json(null);
    assert.deepEqual(await loadFunctionDrafts(scope, controller.signal), []);
    response = new Response("", { status: 403 });
    await assert.rejects(loadFunctionDrafts(scope, controller.signal), /Functions unavailable/);
    response = Response.json({ valid: true });
    assert.deepEqual(await sendFunctionCommand(scope, { operation: "functions.run", name: "echo", source: draft.source, input: { a: 1 } }, controller.signal), { valid: true });
    const call = calls.at(-1)!;
    assert.equal(call.url, "/immutable/admin/functions");
    assert.equal(call.init?.signal, controller.signal);
    assert.equal(new Headers(call.init?.headers).get("authorization"), "Bearer fixture");
    assert.deepEqual(JSON.parse(String(call.init?.body)), { operation: "functions.run", name: "echo", source: draft.source, input: { a: 1 }, project_id: "immutable" });
    response = new Response(null, { status: 204 });
    assert.equal(await sendFunctionCommand(scope, { operation: "functions.permissions", name: "echo" }, controller.signal), null);
    response = new Response("denied", { status: 403 });
    await assert.rejects(sendFunctionCommand(scope, { operation: "functions.publish", name: "echo" }, controller.signal), /denied/);
    response = new Response("", { status: 500 });
    await assert.rejects(sendFunctionCommand(scope, { operation: "functions.publish", name: "echo" }, controller.signal), /500/);
    response = new Response("not JSON");
    await assert.rejects(sendFunctionCommand(scope, { operation: "functions.publish", name: "echo" }, controller.signal), /Invalid function API response/);
  } finally {
    globalThis.fetch = original;
  }
});
