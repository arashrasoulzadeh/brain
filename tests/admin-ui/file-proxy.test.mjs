import assert from "node:assert/strict";
import { test } from "node:test";
import { app, stubFetch } from "./dom-harness.mjs";

const { GET, POST } = app("[project]/admin/files/blob/route.ts");
const request = (method, body = undefined) => {
  const url = new URL("http://localhost/dev/admin/files/blob?version=1&key=object");
  const req = new Request(url, { method, body, headers: { authorization: "Bearer fixture", "content-type": "application/octet-stream", "content-length": "3", "x-project-id": "dev", cookie: "private", "x-project-slug": "dev" } });
  return Object.assign(req, { nextUrl: url });
};
const context = project => ({ params: Promise.resolve({ project }) });

test("binary proxy streams an upload and whitelists response headers", async () => {
  const upload = request("POST", "abc");
  stubFetch(async (url, init) => {
    assert.equal(new URL(url).pathname, "/dev/admin/files/blob");
    assert.equal(new URL(url).searchParams.get("key"), "object");
    assert.equal(init.method, "POST");
    assert.equal(init.body, upload.body);
    assert.equal(init.duplex, "half");
    assert.equal(init.redirect, "error");
    assert.equal(init.headers.get("authorization"), "Bearer fixture");
    assert.equal(init.headers.get("content-length"), "3");
    assert.equal(init.headers.get("cookie"), null);
    assert.equal(await new Response(init.body).text(), "abc");
    return new Response("xyz", { status: 201, headers: { "content-type": "application/octet-stream", "content-disposition": "attachment", "content-length": "3", "set-cookie": "private" } });
  });
  const response = await POST(upload, context("dev"));
  assert.equal(response.status, 201);
  assert.equal(await response.text(), "xyz");
  assert.equal(response.headers.get("content-disposition"), "attachment");
  assert.equal(response.headers.get("set-cookie"), null);
  assert.equal(response.headers.get("cache-control"), "no-store");
  assert.equal(response.headers.get("x-content-type-options"), "nosniff");
});

test("binary proxy rejects malformed projects and handles download and upstream failure", async () => {
  stubFetch(() => { throw new Error("must not fetch"); });
  for (const project of ["../other", "_system", ""]) assert.equal((await GET(request("GET"), context(project))).status, 404);
  stubFetch((url, init) => {
    assert.equal(init.method, "GET");
    assert.equal(init.body, undefined);
    assert.equal(init.headers.get("content-length"), null);
    return new Response("denied", { status: 403 });
  });
  const denied = await GET(request("GET"), context("dev"));
  assert.equal(denied.status, 403);
  assert.equal(await denied.text(), "denied");
  stubFetch(() => { throw new Error("upstream unavailable"); });
  const failure = await GET(request("GET"), context("dev"));
  assert.equal(failure.status, 502);
  assert.deepEqual(await failure.json(), { status: "error", message: "File transfer unavailable" });
});
