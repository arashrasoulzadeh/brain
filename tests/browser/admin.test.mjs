import assert from "node:assert/strict";
import { resolve } from "node:path";
import { test } from "node:test";
import { chromium, artifacts, start, visible, waitFor } from "./browser-runtime.mjs";

test("real Chromium operator login, project management, protected navigation and sign-out", { timeout: 240000 }, async () => {
  const base = "http://127.0.0.1:3100";
  const server = start(process.execPath, ["node_modules/next/dist/bin/next", "dev", "--hostname", "127.0.0.1", "--port", "3100"], "admin-nextjs", "admin");
  let browser;
  try {
    await server.ready(base);
    browser = await chromium.launch({ headless: true });
    const page = await browser.newPage({ viewport: { width: 1280, height: 900 } });
    const requests = [];
    let projects = [{ ID: "dev", Active: true }];
    const response = (route, body, status = 200) => route.fulfill({ status, contentType: "application/json", body: JSON.stringify(body) });
    await page.route("**/api/**", async route => {
      const req = route.request(), url = new URL(req.url()), method = req.method();
      const data = req.postData() ? req.postDataJSON() : undefined;
      requests.push({ path: url.pathname, method, data, authorization: req.headers().authorization });
      if (url.pathname === "/api/admin/session" && method === "POST") {
        return data.secret === "fixture-secret" ? response(route, { session: "browser-fixture" }) : response(route, { message: "Invalid operator credentials" }, 403);
      }
      if (req.headers().authorization !== "Bearer browser-fixture") return response(route, { message: "Forbidden" }, 403);
      if (url.pathname === "/api/admin/session" && method === "DELETE") return response(route, { status: "success" });
      if (url.pathname === "/api/admin/projects") return response(route, projects);
      if (url.pathname === "/api/admin/plugins") return response(route, { manifests: [], installations: [] });
      if (["/api/admin/audit", "/api/admin/system", "/api/admin/functions"].includes(url.pathname)) return response(route, []);
      if (url.pathname === "/api/admin/actions" && data.operation === "project.create") {
        projects = [...projects, { ID: data.project_id, Active: true }];
        return response(route, { status: "success", message: "Project created" });
      }
      return response(route, { message: "Unexpected fixture request" }, 400);
    });
    await page.goto(base, { waitUntil: "networkidle" });
    await visible(page.getByRole("heading", { name: "Welcome back" }));
    await page.getByLabel("Operator username").fill("operator");
    await page.getByLabel("Secret", { exact: true }).fill("wrong");
    await page.getByRole("button", { name: "Sign in", exact: true }).click();
    await visible(page.locator(".login .error"));
    await page.getByLabel("Secret", { exact: true }).fill("fixture-secret");
    await page.getByRole("button", { name: "Sign in", exact: true }).click();
    await visible(page.getByRole("button", { name: "Create project", exact: true }));
    await page.getByPlaceholder("New project ID").fill("browser-project");
    await page.getByRole("button", { name: "Create project", exact: true }).click();
    await visible(page.locator(".project-table-row").filter({ hasText: "browser-project" }));
    await page.locator("#selected-project").selectOption("browser-project");
    await waitFor(() => page.evaluate(() => sessionStorage.getItem("brain-admin-project") === "browser-project"), "project selection was not persisted");
    await page.screenshot({ path: resolve(artifacts, "admin-projects.png"), fullPage: true });
    assert.ok(requests.some(req => req.data?.operation === "project.create" && req.authorization === "Bearer browser-fixture"));
    await page.goto(base + "/system", { waitUntil: "networkidle" });
    await visible(page.getByText("This page is unavailable until its plugin is ready and enabled.", { exact: true }));
    await page.goto(base, { waitUntil: "networkidle" });
    await visible(page.getByRole("button", { name: "Sign out", exact: true }));
    await page.getByRole("button", { name: "Sign out", exact: true }).click();
    await visible(page.getByRole("heading", { name: "Welcome back" }));
    assert.equal(await page.evaluate(() => sessionStorage.getItem("brain-admin-session")), null);
    assert.ok(requests.some(req => req.method === "DELETE" && req.authorization === "Bearer browser-fixture"));
  } finally { await browser?.close(); server.close(); }
});
