import assert from "node:assert/strict";
import { test } from "node:test";
import { app, mount, find, button, click, change, submit, text, session, stubFetch, json } from "./dom-harness.mjs";

const AdminPage = app("page.tsx").default;

test("operator login, project creation, project selection and sign-out work through the actual main page", async () => {
  const calls = [];
  let projects = [{ ID: "dev", Active: true }, { ID: "other", Active: true }];
  let loginDenied = true;
  stubFetch((url, init) => {
    const body = init.body ? JSON.parse(init.body) : undefined;
    calls.push({ url, method: init.method, body, authorization: init.headers?.authorization });
    if (url === "/api/admin/session" && init.method === "POST") return loginDenied ? json({ message: "Invalid operator login" }, 403) : json({ session: "fixture-token" });
    if (url === "/api/admin/session" && init.method === "DELETE") return json({ status: "success" });
    if (url === "/api/admin/projects") return json(projects);
    if (url.startsWith("/api/admin/plugins")) return json({ manifests: [], installations: [] });
    if (url === "/api/admin/audit" || url === "/api/admin/system" || url.startsWith("/api/admin/functions")) return json([]);
    if (url === "/api/admin/actions" && body.operation === "project.create") {
      projects = [...projects, { ID: body.project_id, Active: true }];
      return json({ status: "success" });
    }
    throw new Error("unexpected request " + url);
  });
  await mount(AdminPage);
  await click(button("فارسی"));
  assert.equal(find("main").dir, "rtl");
  await click(button("English"));
  assert.equal(find("main").dir, "ltr");
  await change(find("form input:not([type])"), "operator");
  await change(find('input[type="password"]'), "wrong");
  await submit(find("form"));
  assert.match(text(), /Invalid operator login/);
  assert.deepEqual(calls.at(-1).body, { username: "operator", secret: "wrong" });
  loginDenied = false;
  await change(find('input[type="password"]'), "secret");
  await submit(find("form"));
  assert.equal(window.sessionStorage.getItem("brain-admin-session"), "fixture-token");
  assert.equal(find("#selected-project").value, "dev");
  await change(find(".create input"), "new-project");
  await submit(find(".create"));
  assert.ok(calls.some(call => call.body?.operation === "project.create" && call.body.project_id === "new-project" && call.authorization === "Bearer fixture-token"));
  assert.match(text(), /new-project/);
  await change(find("#selected-project"), "other");
  assert.equal(window.sessionStorage.getItem("brain-admin-project"), "other");
  assert.ok(calls.some(call => call.url === "/api/admin/plugins?project_id=other"));
  await click(button("Sign out"));
  assert.equal(window.sessionStorage.getItem("brain-admin-session"), null);
  assert.equal(window.sessionStorage.getItem("brain-admin-project"), null);
  assert.ok(calls.some(call => call.method === "DELETE" && call.authorization === "Bearer fixture-token"));
  assert.ok(find('input[type="password"]'));
});

test("restored session with a failed project request is removed, and disabled route stays unavailable", async () => {
  session();
  stubFetch(url => url === "/api/admin/projects" ? json({ message: "Expired" }, 403) : json([]));
  await mount(AdminPage);
  assert.match(text(), /session expired/);
  assert.equal(window.sessionStorage.getItem("brain-admin-session"), null);
  session();
  stubFetch(url => {
    if (url === "/api/admin/projects") return json([{ ID: "dev", Active: true }]);
    if (url.startsWith("/api/admin/plugins")) return json({ manifests: [], installations: [] });
    return json([]);
  });
  await mount(AdminPage, {}, { pathname: "/functions" });
  assert.match(text(), /This page is unavailable until its plugin is ready and enabled/);
});
