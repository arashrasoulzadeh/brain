import assert from "node:assert/strict";
import { test } from "node:test";
import { adminRequestIdentity, adminSessionRejected } from "../../admin-nextjs/app/lib/admin-session-response.js";

const origin = "http://localhost:3100";
const headers = new Headers({ authorization: "Bearer token" });

test("only authenticated same-origin admin requests can invalidate the session", () => {
  for (const path of ["/_system/admin/projects", "/alpha/admin/actions", "/alpha/admin/function-logs"]) {
    assert.deepEqual(adminRequestIdentity(path, headers, origin), { path, token: "token" });
  }
  assert.deepEqual(adminRequestIdentity(new URL(origin + "/alpha/admin/plugins"), headers, origin), { path: "/alpha/admin/plugins", token: "token" });
  assert.deepEqual(adminRequestIdentity(new Request(origin + "/_system/admin/audit"), headers, origin), { path: "/_system/admin/audit", token: "token" });
  for (const path of ["https://elsewhere.test/alpha/admin/actions", "/alpha/files/upload", "/alpha/ui/files", "/_next/static/main.js"]) {
    assert.equal(adminRequestIdentity(path, headers, origin), undefined);
  }
  for (const authorization of ["", "Bearer", "Basic token", "Bearer "]) {
    assert.equal(adminRequestIdentity("/_system/admin/session", new Headers({ authorization }), origin), undefined);
  }
});

test("401 expires a session; 403 requires proof from the operator project endpoint", async () => {
  const identity = { path: "/alpha/admin/actions", token: "token" };
  let calls = 0;
  let status = 200;
  const request: typeof fetch = async (input, init) => {
    calls++;
    assert.equal(input, "/_system/admin/projects");
    assert.equal(new Headers(init?.headers).get("authorization"), "Bearer token");
    assert.equal(init?.cache, "no-store");
    return new Response(null, { status });
  };
  assert.equal(await adminSessionRejected(identity, 401, request), true);
  assert.equal(await adminSessionRejected(identity, 500, request), false);
  assert.equal(await adminSessionRejected({ ...identity, path: "/_system/admin/projects" }, 403, request), true);
  assert.equal(calls, 0);
  for (const [response, expected] of [[200, false], [401, true], [403, true], [503, false]] as const) {
    status = response;
    assert.equal(await adminSessionRejected(identity, 403, request), expected);
  }
  assert.equal(calls, 4);
  assert.equal(await adminSessionRejected(identity, 403, async () => { throw new Error("offline"); }), false);
});
