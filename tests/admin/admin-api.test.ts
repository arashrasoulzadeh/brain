import assert from "node:assert/strict";
import { test } from "node:test";
import { adminHeaders, revokeAdminSession, projectAdminURL, systemAdminURL, projectUIURL, systemUIURL, adminUIContext, preferredAdminProject } from "../../admin-nextjs/app/lib/admin-api.js";

test("admin requests use one bearer header without dropping custom headers", () => {
  const headers = adminHeaders("token", { "content-type": "application/json", "x-brain-quiet": "1" });
  assert.equal(headers.authorization, "Bearer token");
  assert.equal(headers["content-type"], "application/json");
  assert.equal(headers["x-brain-quiet"], "1");
});

test("shared sign-out transport revokes the server session", async () => {
  const calls: { url: string; init?: RequestInit }[] = [];
  const request: typeof fetch = async (input, init) => {
    calls.push({ url: String(input), init });
    return new Response(null, { status: 204 });
  };
  await revokeAdminSession("", request);
  await revokeAdminSession("operator", request);
  assert.equal(calls.length, 1);
  assert.equal(calls[0].url, "/_system/admin/session");
  assert.equal(calls[0].init?.method, "DELETE");
  assert.equal(new Headers(calls[0].init?.headers).get("authorization"), "Bearer operator");
});

test("project admin URLs require an explicit project ID or code", () => {
  assert.equal(projectAdminURL("game-one", "plugins"), "/game-one/admin/plugins");
  assert.equal(projectAdminURL("game-two", "actions"), "/game-two/admin/actions");
  assert.equal(projectAdminURL("game-one", "functions"), "/game-one/admin/functions");
  assert.equal(systemAdminURL("projects"), "/_system/admin/projects");
  for (const project of ["", "_system", "../game", "game/other", "game%2Fother"]) {
    assert.throws(() => projectAdminURL(project, "actions"), /invalid project API URL/);
  }
  for (const resource of ["", "/actions", "plugins//one", "../actions"]) {
    assert.throws(() => projectAdminURL("game-one", resource), /invalid project API URL/);
  }
});

test("admin navigation resolves the project in its URL rather than saved selection", () => {
  assert.equal(systemUIURL(), "/_system/ui");
  assert.equal(systemUIURL("projects"), "/_system/ui/projects");
  assert.equal(projectUIURL("game-code", "functions", "hello world"), "/game-code/ui/functions/hello%20world");
  assert.deepEqual(adminUIContext("/game-code/ui/functions/hello"), { namespace: "game-code", page: "functions" });
  assert.deepEqual(adminUIContext("/_system/ui/projects"), { page: "projects" });
  assert.deepEqual(adminUIContext("/game-code/ui"), { namespace: "game-code", page: "plugins" });
  for (const project of ["", "_system", "../game", "game/other"]) assert.throws(() => projectUIURL(project, "functions"));
  for (const segment of ["", ".", "..", "a/b", "a\\b"]) assert.throws(() => projectUIURL("game", segment));
  const projects = [{ ID: "first", Slug: "first-code", Active: true }, { ID: "second", Slug: "second-code", Active: true }, { ID: "inactive", Slug: "off", Active: false }];
  assert.equal(preferredAdminProject(projects, "second-code")?.ID, "second");
  assert.equal(preferredAdminProject(projects, "first")?.ID, "first");
  assert.equal(preferredAdminProject(projects, "missing"), undefined);
  assert.equal(preferredAdminProject(projects, "off"), undefined);
});
