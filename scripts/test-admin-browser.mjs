import assert from "node:assert/strict";
import { mkdir } from "node:fs/promises";
import { createRequire } from "node:module";
import { resolve } from "node:path";

const require = createRequire(resolve(process.env.BRAIN_BROWSER_DEPS ?? ".", "package.json"));
const { chromium } = require("playwright");
const origin = process.env.BRAIN_ADMIN_ORIGIN ?? "http://127.0.0.1:3100";
const output = resolve(process.env.BRAIN_BROWSER_OUTPUT ?? ".local/admin-browser");
const projects = [{ ID: "alpha", Active: true }, { ID: "beta", Active: true }];
const manifest = {
  ID: "files", Title: "Files", Description: "Project-scoped object metadata.", Version: "0.1.0",
  Capabilities: ["files.metadata", "files.publish"],
  Routes: [{ Operation: "files.metadata", Capability: "files.metadata" }, { Operation: "files.publish", Capability: "files.publish" }, { Operation: "files.model.list", Capability: "files.metadata" }, { Operation: "files.model.read", Capability: "files.metadata" }],
  Admin: { version: 1, pages: [{
    id: "operations", title: "File metadata",
    inventory: { kind: "file_keys", capability: "files.metadata", title: "Files", select_field: "key" },
    actions: [
      { title: "Inspect file", operation: "files.metadata", read_only: true, fields: [{ name: "key", type: "string", required: true }], result: { type: "detail" } },
      { title: "Publish file", operation: "files.publish", fields: [{ name: "key", type: "string", required: true }] },
    ],
  }, {
    id: "model-record", title: "Model records", model: { name: "record", schema: { version: 1, fields: { name: "string" } } },
    actions: [
      { title: "list", operation: "files.model.list", read_only: true, model_action: "list", fields: [{ name: "after", type: "string" }] },
      { title: "read", operation: "files.model.read", read_only: true, model_action: "read", fields: [{ name: "id", type: "string", required: true }] },
    ],
  }] },
};
const commands = [];
const runtimeErrors = [];
const browser = await chromium.launch({ headless: true, executablePath: process.env.BRAIN_BROWSER_EXECUTABLE || undefined });
await mkdir(output, { recursive: true });

async function fixture(viewport, project = "alpha") {
  const context = await browser.newContext({ viewport });
  await context.addInitScript(({ project }) => {
    if (!sessionStorage.getItem("brain-admin-session") && !sessionStorage.getItem("brain-browser-initialized")) {
      sessionStorage.setItem("brain-admin-session", "browser-fixture");
      sessionStorage.setItem("brain-admin-project", project);
      sessionStorage.setItem("brain-browser-initialized", "yes");
    }
  }, { project });
  const page = await context.newPage();
  let allowed = true;
  let modelAllowed = true;
  let sessionStatus = 0;
  let betaInstalled = false;
  let betaEnabled = false;
  page.on("pageerror", error => runtimeErrors.push(error.message));
  await page.route("**/*/admin/**", async route => {
    const request = route.request();
    const url = new URL(request.url());
    if (url.pathname === "/_system/admin/session" && request.method() === "POST") {
      if (request.postDataJSON().secret !== "valid-fixture") return route.fulfill({ status: 401, json: { message: "Invalid credentials" } });
      sessionStatus = 0;
      return route.fulfill({ json: { session: "browser-fixture" } });
    }
    if (request.headers().authorization !== "Bearer browser-fixture") {
      return route.fulfill({ status: 403, body: "Forbidden" });
    }
    if (sessionStatus) return route.fulfill({ status: sessionStatus, body: "Operator session expired" });
    const json = body => route.fulfill({ json: body });
    const namespace = url.pathname.split("/")[1];
    const endpoint = url.pathname.replace(/^\/[A-Za-z0-9_-]+\/admin\//, "/admin/");
    switch (endpoint) {
      case "/admin/projects": return json(projects);
      case "/admin/plugins": {
        const id = url.searchParams.get("project_id") ?? (namespace === "_system" ? "" : namespace);
        return json({
          manifests: [manifest],
          installations: id === "alpha" || (id === "beta" && betaInstalled) ? [{ ProjectID: id, PluginID: "files", Version: manifest.Version,
            Enabled: id === "alpha" || betaEnabled, Ready: true, WorkerReady: id === "alpha" || betaEnabled, Draining: false, Grants: manifest.Capabilities }] : [],
        });
      }
      case "/admin/routes": return json(namespace === "alpha" ? [{ plugin_id: "files", version: manifest.Version, method: "POST", path: "/alpha/files/metadata", function: "metadata", capability: "files.metadata", operation: "files.metadata", auth: "operator" }] : []);
      case "/admin/audit":
      case "/admin/system":
      case "/admin/functions": return json([]);
      case "/admin/session": return route.fulfill({ status: 204 });
      case "/admin/actions": {
        const body = request.postDataJSON();
        commands.push(body);
        if (namespace === "beta") {
          assert.equal(body.project_id, "beta");
          assert.equal(body.plugin_id, "files");
          assert.equal(body.version, manifest.Version);
          if (body.operation === "plugin.install.activate") {
            assert.equal(body.confirmed, true);
            assert.deepEqual(body.capabilities, manifest.Capabilities);
            betaInstalled = true; betaEnabled = true;
            return json({ status: "success" });
          }
          if (body.operation === "plugin.disable.preview") return json({ Revision: 1, Dependents: [] });
          if (body.operation === "plugin.disable.apply") { assert.equal(body.confirmed, true); betaEnabled = false; return json({ status: "success" }); }
          if (body.operation === "plugin.enable") { betaEnabled = true; return json({ status: "success" }); }
          return route.fulfill({ status: 400, body: "Unsupported beta command" });
        }
        assert.equal(namespace, "alpha");
        assert.equal(body.project_id, "alpha");
        assert.equal(body.plugin_id, "files");
        assert.equal(body.version, manifest.Version);
        if (body.page_id === "model-record") {
          assert.equal(body.target_operation, "files.model.list");
          if (!modelAllowed) return route.fulfill({ status: 403, body: "Model permission revoked" });
          if (body.operation === "plugin.admin.revision") return json({ revision: 0 });
          assert.equal(body.operation, "plugin.admin.query");
          return json({ Success: true, Payload: Buffer.from(JSON.stringify([{ id: "record-one", revision: "fixture-revision", fields: { name: "private-name" } }])).toString("base64") });
        }
        if (body.operation === "plugin.invoke") {
          assert.equal(body.target_operation, "files.metadata");
          assert.deepEqual(body.payload, { key: "sample.txt" });
          return json({ Success: false, ErrorCode: "fixture-failure", Payload: Buffer.from(JSON.stringify({ detail: "résultat" })).toString("base64") });
        }
        assert.equal(body.page_id, "operations");
        if (body.operation === "plugin.admin.inventory") {
          return json({ rows: allowed ? [{ key: "sample.txt", size_bytes: 12, published: false }] : [] });
        }
        if (body.operation === "plugin.admin.revision") {
          assert.equal(body.target_operation, "files.metadata");
          assert.deepEqual(body.payload, { key: "sample.txt" });
          return allowed ? json({ revision: 0 }) : route.fulfill({ status: 403, body: "Resource permission revoked" });
        }
        if (body.operation === "plugin.admin.query") {
          assert.equal(body.target_operation, "files.metadata");
          assert.equal(body.confirmed, false);
          assert.deepEqual(body.payload, { key: "sample.txt" });
        } else if (body.operation === "plugin.admin.invoke") {
          assert.equal(body.target_operation, "files.publish");
          assert.equal(body.confirmed, true);
          assert.deepEqual(body.payload, { key: "sample.txt" });
        } else {
          return route.fulfill({ status: 400, body: "Unsupported command" });
        }
        return json({ Success: true, Payload: Buffer.from(JSON.stringify({ key: "sample.txt", status: "fixture-result" })).toString("base64") });
      }
      default: return route.fulfill({ status: 404, body: "Unknown fixture endpoint" });
    }
  });
  return { page, context, setAllowed(value) { allowed = value; }, setModelAllowed(value) { modelAllowed = value; }, setSessionStatus(value) { sessionStatus = value; } };
}

async function visible(locator) {
  await locator.waitFor({ state: "visible", timeout: 15000 });
}
async function menuSearch(page, query) {
  await page.getByRole("textbox", { name: "Filter navigation" }).fill(query);
}
async function screenshot(page, name) {
  await page.screenshot({ path: resolve(output, name + ".png"), fullPage: true, animations: "disabled" });
}

try {
  const desktop = await fixture({ width: 1440, height: 900 });
  const page = desktop.page;
  await page.goto(origin + "/alpha/ui/plugins");
  await visible(page.getByRole("heading", { name: "Plugins", level: 1 }));
  await visible(page.getByRole("link", { name: "Files", exact: true }));
  assert.equal(await page.locator("aside.sidebar").count(), 1);
  assert.equal(Math.round((await page.locator("aside.sidebar").boundingBox()).width), 224);
  assert.equal(await page.getByRole("link", { name: "Plugins", exact: true }).getAttribute("aria-current"), "page");
  await screenshot(page, "desktop-plugins");

  await menuSearch(page, "file metadata");
  const dynamicLink = page.getByRole("navigation", { name: "Admin navigation" }).getByRole("link", { name: /File metadata/ });
  await visible(dynamicLink);
  assert.equal(await dynamicLink.getAttribute("href"), "/alpha/ui/extensions/files/0.1.0/operations");
  await dynamicLink.click();
  await visible(page.getByRole("heading", { name: "File metadata", level: 1 }));
  assert.ok(new URL(page.url()).pathname.startsWith("/alpha/ui/"));
  await visible(page.getByText("sample.txt", { exact: true }));
  assert.equal(await page.getByRole("link", { name: "Manage files in alpha" }).getAttribute("href"), "/alpha/ui/plugins/files%400.1.0");
  assert.equal(await page.getByRole("dialog").count(), 0);
  await page.getByRole("button", { name: "Select", exact: true }).click();
  await visible(page.getByRole("dialog", { name: "Inspect file" }));
  assert.equal(await page.getByRole("textbox", { name: /key/ }).inputValue(), "sample.txt");
  await page.getByRole("button", { name: "Query", exact: true }).click();
  await visible(page.getByText("fixture-result", { exact: true }));
  await screenshot(page, "desktop-plugin-inventory");
  // Revocation has no observation event. Both the inventory and the visible
  // query result must still disappear on their next authorization check.
  desktop.setAllowed(false);
  await visible(page.getByText("No resources on this page.", { exact: true }));
  await visible(page.getByRole("alert").filter({ hasText: "Resource permission revoked" }));
  assert.equal(await page.evaluate(() => sessionStorage.getItem("brain-admin-session")), "browser-fixture");
  assert.equal(await page.locator("aside.sidebar").count(), 1);
  assert.equal(await page.getByText("fixture-result", { exact: true }).count(), 0);
  desktop.setAllowed(true);
  await visible(page.getByText("sample.txt", { exact: true }));

  // A mutation must not run when the operator cancels confirmation.
  await page.getByLabel("Operation").selectOption("files.publish");
  await page.getByRole("textbox", { name: /key/ }).fill("sample.txt");
  await page.getByRole("button", { name: "Run action" }).click();
  await page.getByRole("button", { name: "Cancel confirmation" }).click();
  assert.equal(commands.filter(command => command.operation === "plugin.admin.invoke").length, 0);
  await page.getByRole("button", { name: "Run action" }).click();
  await page.getByRole("button", { name: "Confirm action" }).click();
  await visible(page.getByText("fixture-result", { exact: false }).last());
  assert.equal(commands.filter(command => command.operation === "plugin.admin.invoke").length, 1);

  await page.getByRole("button", { name: "Close dialog" }).click();
  assert.equal(await page.getByRole("dialog").count(), 0);
  await menuSearch(page, "Model records");
  await page.getByRole("navigation", { name: "Admin navigation" }).getByRole("link", { name: /Model records/ }).click();
  await visible(page.getByText("record-one", { exact: true }));
  await page.getByRole("button", { name: "Select", exact: true }).click();
  assert.equal(await page.getByLabel("Record ID").inputValue(), "record-one");
  assert.equal(await page.getByLabel("name", { exact: true }).inputValue(), "private-name");
  desktop.setModelAllowed(false);
  await visible(page.getByRole("status").filter({ hasText: "Model permission revoked" }));
  assert.equal(await page.getByText("record-one", { exact: true }).count(), 0);
  assert.equal(await page.getByLabel("Record ID").inputValue(), "");
  assert.equal(await page.getByLabel("name", { exact: true }).inputValue(), "");
  await page.getByRole("button", { name: "Close dialog" }).click();

  await page.getByRole("navigation", { name: "Admin navigation" }).getByRole("link", { name: "Projects", exact: true }).click();
  await visible(page.getByRole("heading", { name: "Projects", level: 1 }));
  await page.locator("#selected-project").selectOption("beta");
  await visible(page.locator(".workspace-picker").getByText("beta", { exact: true }));
  await menuSearch(page, "File metadata");
  await visible(page.getByText("No matching pages.", { exact: true }));
  await page.evaluate(() => sessionStorage.setItem("brain-admin-project", "alpha"));
  await page.goto(origin + "/beta/ui/extensions/files/0.1.0/operations");
  assert.equal(new URL(page.url()).pathname, "/beta/ui/extensions/files/0.1.0/operations");
  await visible(page.getByRole("alert").filter({ hasText: "Page unavailable or permission revoked" }));
  await desktop.context.close();

  // Confirmation is read-only until accepted; sidebar refreshes without navigation.
  const installation = await fixture({ width: 1440, height: 900 }, "beta");
  const target = installation.page;
  await target.goto(origin + "/beta/ui/plugins");
  const nav = target.getByRole("navigation", { name: "Admin navigation" });
  await visible(target.getByRole("button", { name: "Install", exact: true }));
  await menuSearch(target, "File metadata");
  assert.equal(await nav.getByRole("link", { name: /File metadata/ }).count(), 0);
  const beforeInstall = commands.length;
  await target.getByRole("button", { name: "Install", exact: true }).click();
  const consent = target.getByRole("dialog", { name: "Install Files" });
  await visible(consent);
  for (const capability of manifest.Capabilities) await visible(consent.getByText(capability, { exact: true }));
  await consent.getByRole("button", { name: "Cancel", exact: true }).click();
  assert.equal(commands.length, beforeInstall);
  await target.getByRole("button", { name: "Install", exact: true }).click();
  await consent.getByRole("button", { name: "Install and enable" }).click();
  await visible(consent.getByRole("status").filter({ hasText: "Plugin installed and enabled" }));
  await consent.getByRole("button", { name: "Done" }).click();
  await visible(nav.getByRole("link", { name: /File metadata/ }));
  assert.equal(new URL(target.url()).pathname, "/beta/ui/plugins");
  await target.getByRole("link", { name: "View files" }).click();
  await visible(target.getByRole("heading", { name: "Project settings" }));
  await menuSearch(target, "File metadata");
  await visible(nav.getByRole("link", { name: /File metadata/ }));
  await target.getByRole("button", { name: "Disable", exact: true }).click();
  await target.getByRole("button", { name: "Confirm action" }).click();
  await visible(target.getByRole("status").filter({ hasText: "Action completed" }));
  await target.getByRole("button", { name: "Close dialog" }).click();
  await nav.getByRole("link", { name: /File metadata/ }).waitFor({ state: "detached" });
  await target.getByRole("button", { name: "Enable", exact: true }).click();
  await target.getByRole("button", { name: "Confirm action" }).click();
  await visible(target.getByRole("status").filter({ hasText: "Action completed" }));
  await target.getByRole("button", { name: "Close dialog" }).click();
  await visible(nav.getByRole("link", { name: /File metadata/ }));
  await installation.context.close();

  // Page access failure stays in the authenticated shell. The route
  // playground uses the exact discovered version and does not replace input.
  const review = await fixture({ width: 1440, height: 900 });
  const reviewed = review.page;
  await reviewed.goto(origin + "/alpha/ui/functions");
  await visible(reviewed.getByRole("alert").filter({ hasText: "This plugin is disabled" }));
  assert.equal(new URL(reviewed.url()).pathname, "/alpha/ui/functions");
  assert.equal(await reviewed.locator("aside.sidebar").count(), 1);
  assert.equal(await reviewed.evaluate(() => sessionStorage.getItem("brain-admin-session")), "browser-fixture");
  await reviewed.goto(origin + "/alpha/ui/routes");
  await reviewed.getByLabel("Search routes").fill("not-a-route");
  await visible(reviewed.getByText("No matching routes.", { exact: true }));
  await reviewed.getByLabel("Search routes").fill("files.metadata");
  await reviewed.getByRole("button", { name: "Use in playground", exact: true }).click();
  assert.equal(await reviewed.getByLabel("Select a route").inputValue(), "files:" + manifest.Version + ":files.metadata");
  await reviewed.getByLabel("JSON payload", { exact: true }).fill('{"key":"sample.txt"}');
  const beforeInvoke = commands.length;
  await reviewed.getByLabel("JSON payload", { exact: true }).fill("{");
  await reviewed.getByRole("button", { name: "Send", exact: true }).click();
  await visible(reviewed.getByText("Invalid JSON payload", { exact: true }));
  assert.equal(commands.length, beforeInvoke);
  await reviewed.getByLabel("JSON payload", { exact: true }).fill('{"key":"sample.txt"}');
  await reviewed.getByRole("button", { name: "Send", exact: true }).click();
  const invocation = reviewed.getByRole("dialog", { name: "Invoke plugin operation" });
  await visible(invocation);
  assert.equal(commands.length, beforeInvoke);
  await invocation.getByRole("button", { name: "Confirm invocation" }).click();
  await visible(invocation.getByRole("status").filter({ hasText: "Plugin error: fixture-failure" }));
  await visible(invocation.getByText(/résultat/));
  assert.equal(await invocation.getByRole("button", { name: "Confirm invocation" }).count(), 0);
  await invocation.getByRole("button", { name: "Close dialog" }).click();
  assert.equal(await reviewed.getByLabel("JSON payload", { exact: true }).inputValue(), '{"key":"sample.txt"}');
  await reviewed.goto(origin + "/_system/ui/projects");
  await visible(reviewed.getByRole("heading", { name: "Projects", level: 1 }));
  const beforeProject = commands.length;
  await reviewed.getByRole("button", { name: "Configure", exact: true }).first().click();
  await visible(reviewed.getByRole("dialog", { name: "Configure project" }));
  await reviewed.getByLabel("Display name", { exact: true }).fill("Cancelled edit");
  await reviewed.keyboard.press("Escape");
  assert.equal(commands.length, beforeProject);
  assert.equal(await reviewed.getByRole("dialog").count(), 0);
  await review.context.close();

  // The URL must win over the saved selection in a fresh browser context.
  const mobile = await fixture({ width: 390, height: 844 }, "beta");
  await mobile.page.goto(origin + "/alpha/ui/plugins");
  await visible(mobile.page.getByRole("heading", { name: "Plugins", level: 1 }));
  const toggle = mobile.page.getByRole("button", { name: "Toggle navigation" });
  await visible(toggle);
  assert.equal(await toggle.getAttribute("aria-expanded"), "false");
  await toggle.click();
  await visible(mobile.page.locator(".workspace-picker").getByText("alpha", { exact: true }));
  await menuSearch(mobile.page, "file metadata");
  await visible(mobile.page.getByRole("navigation", { name: "Admin navigation" }).getByRole("link", { name: /File metadata/ }));
  await screenshot(mobile.page, "mobile-navigation");
  await mobile.page.getByRole("navigation", { name: "Admin navigation" }).getByRole("link", { name: /File metadata/ }).click();
  await visible(mobile.page.getByRole("heading", { name: "File metadata", level: 1 }));
  assert.equal(await toggle.getAttribute("aria-expanded"), "false");
  assert.ok(await mobile.page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth + 1), "mobile page must not overflow horizontally");
  await screenshot(mobile.page, "mobile-plugin");
  await mobile.page.getByRole("button", { name: "Inspect file", exact: true }).click();
  await visible(mobile.page.getByRole("dialog", { name: "Inspect file" }));
  assert.ok(await mobile.page.getByRole("dialog").evaluate(dialog => dialog.getBoundingClientRect().width <= window.innerWidth));
  await screenshot(mobile.page, "mobile-action-dialog");
  await mobile.page.keyboard.press("Escape");
  assert.equal(await mobile.page.getByRole("dialog").count(), 0);
  await mobile.context.close();

  // A 401 from quiet polling must clear the session and remove the entire
  // protected shell, rather than nesting the login form inside its content.
  for (const status of [401, 403]) {
    const expired = await fixture({ width: 1440, height: 900 });
    const target = expired.page;
    await target.goto(origin + "/alpha/ui/plugins");
    await visible(target.getByRole("link", { name: "Files", exact: true }));
    expired.setSessionStatus(status);
    await target.evaluate(() => { void fetch("/alpha/admin/functions", {
      headers: { authorization: "Bearer browser-fixture", "x-brain-quiet": "1" },
    }).catch(() => undefined); });
    await target.waitForURL(origin + "/_system/ui");
    await visible(target.getByRole("heading", { name: "Welcome back" }));
    assert.equal(await target.locator("aside.sidebar").count(), 0);
    assert.equal(await target.locator("#admin-content").count(), 0);
    assert.equal(await target.evaluate(() => sessionStorage.getItem("brain-admin-session")), null);
    assert.equal(await target.evaluate(() => sessionStorage.getItem("brain-admin-project")), null);
    await screenshot(target, "expired-session-" + status);

    // Direct entry to a protected URL without a token also uses standalone login.
    await target.goto(origin + "/alpha/ui/plugins");
    await target.waitForURL(origin + "/_system/ui");
    await visible(target.getByRole("heading", { name: "Welcome back" }));
    assert.equal(await target.locator("aside.sidebar").count(), 0);
    await target.getByLabel("Operator username").fill("operator");
    await target.getByLabel("Secret", { exact: true }).fill("wrong");
    await target.getByRole("button", { name: "Sign in", exact: true }).click();
    await visible(target.getByText("Invalid credentials", { exact: true }));
    assert.equal(new URL(target.url()).pathname, "/_system/ui");
    assert.equal(await target.locator("aside.sidebar").count(), 0);
    await target.getByLabel("Secret", { exact: true }).fill("valid-fixture");
    await target.getByRole("button", { name: "Sign in", exact: true }).click();
    await visible(target.getByRole("heading", { name: "Projects", level: 1 }));
    assert.equal(await target.locator("aside.sidebar").count(), 1);
    assert.equal(await target.evaluate(() => sessionStorage.getItem("brain-admin-session")), "browser-fixture");
    await expired.context.close();
  }

  assert.deepEqual(runtimeErrors, [], "browser runtime errors");
  console.log("Admin browser smoke passed: desktop, mobile, permissions, project switching and action confirmation.");
} finally {
  await browser.close();
}
