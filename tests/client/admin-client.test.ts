import assert from "node:assert/strict";
import { test } from "node:test";
import { adminAction, adminLogin, projectHTTPURL } from "../../clients/javascript/src/index.js";

const calls: RequestInit[] = [];
const urls: string[] = [];
const request = async (_url: string | URL | Request, init?: RequestInit) => {
  calls.push(init ?? {});
  urls.push(String(_url));
  return new Response(JSON.stringify({ session: "operator-session" }), { status: 200 });
};
const login = await adminLogin("http://localhost:18081", "admin", "secret", request);
if (login.session !== "operator-session") throw new Error("admin login client failed");
await adminAction(
  "http://localhost:18081",
  login.session,
  { operation: "plugin.enable", project_id: "dev" },
  request,
);
if (
  calls.length !== 2 ||
  new Headers(calls[1]?.headers).get("authorization") !== "Bearer operator-session"
)
  throw new Error("admin action client failed");

if (urls[0] !== "http://localhost:18081/_system/admin/session" ||
    urls[1] !== "http://localhost:18081/dev/admin/actions") {
  throw new Error("operator URL must be project or system scoped");
}

await adminAction("http://localhost:18081", login.session, { operation: "project.create", project_id: "new-project" }, request);
await adminAction("http://localhost:18081", login.session, { operation: "project.activate", project_id: "dev" }, request);
await adminAction("http://localhost:18081", login.session, { operation: "plugin.readiness", plugin_id: "rooms" }, request);
if (urls.slice(2).some(url => url !== "http://localhost:18081/_system/admin/actions"))
  throw new Error("control-plane operation must use system namespace");
let missingProject = false;
try {
  await adminAction("http://localhost:18081", login.session, { operation: "plugin.enable" }, request);
} catch { missingProject = true; }
if (!missingProject || urls.length !== 5) throw new Error("projectless plugin command was sent");

test("project HTTP URL accepts only canonical project namespaces and resources", () => {
  assert.equal(projectHTTPURL("https://example.test:8443/", "Game_1-prod", "/auth/session"), "https://example.test:8443/Game_1-prod/auth/session");
  assert.equal(projectHTTPURL("http://localhost:18081", "dev", "/ws"), "http://localhost:18081/dev/ws");

  for (const project of ["", "_system", "_private", "-dev", "a/b", "a.b", "a%2fb", " a"]) {
    assert.throws(() => projectHTTPURL("https://example.test", project, "/ws"), /invalid project namespace/);
  }
  for (const resource of ["ws", "/", "//ws", "/ws/", "/a/../ws", "/a%2fb", "/ws?x=1", "/ws#fragment"]) {
    assert.throws(() => projectHTTPURL("https://example.test", "dev", resource), /invalid project HTTP resource/);
  }
  for (const origin of [
    "ftp://example.test",
    "https://alice@example.test",
    "https://:secret@example.test",
    "https://example.test/path",
    "https://example.test/?x=1",
    "https://example.test/#fragment",
  ]) {
    assert.throws(() => projectHTTPURL(origin, "dev", "/ws"), /invalid HTTP API origin/);
  }
  assert.throws(() => projectHTTPURL("not a URL", "dev", "/ws"), TypeError);
});

test("project admin action rejects malformed namespace before sending a request", async () => {
  const unexpectedRequest: typeof fetch = async () => {
    assert.fail("invalid project namespace must not reach transport");
  };
  await assert.rejects(
    adminAction("https://example.test", "token", { operation: "plugin.enable", project_id: "_system" }, unexpectedRequest),
    /invalid project namespace/,
  );
  await assert.rejects(
    adminAction("https://example.test/path", "token", { operation: "plugin.enable", project_id: "dev" }, unexpectedRequest),
    /invalid HTTP API origin/,
  );
});
