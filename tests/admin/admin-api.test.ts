import assert from "node:assert/strict";
import { test } from "node:test";
import { adminHeaders, revokeAdminSession } from "../../admin-nextjs/app/lib/admin-api.js";

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
  assert.equal(calls[0].url, "/api/admin/session");
  assert.equal(calls[0].init?.method, "DELETE");
  assert.equal(new Headers(calls[0].init?.headers).get("authorization"), "Bearer operator");
});
