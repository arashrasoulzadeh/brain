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
  page.on("pageerror", error => runtimeErrors.push(error.message));
  await page.route("**/*/admin/**", async route => {
    const request = route.request();
    const url = new URL(request.url());
    if (request.headers().authorization !== "Bearer browser-fixture") {
      return route.fulfill({ status: 403, body: "Forbidden" });
    }
    const json = body => route.fulfill({ json: body });
    const namespace = url.pathname.split("/")[1];
    const endpoint = url.pathname.replace(/^\/[A-Za-z0-9_-]+\/admin\//, "/admin/");
    switch (endpoint) {
      case "/admin/projects": return json(projects);
      case "/admin/plugins": {
        const id = url.searchParams.get("project_id") ?? (namespace === "_system" ? "" : namespace);
        return json({
          manifests: [manifest],
          installations: id === "alpha" ? [{ ProjectID: id, PluginID: "files", Version: manifest.Version,
            Enabled: true, Ready: true, Draining: false, Grants: manifest.Capabilities }] : [],
        });
      }
      case "/admin/audit":
      case "/admin/system":
      case "/admin/functions": return json([]);
      case "/admin/session": return route.fulfill({ status: 204 });
      case "/admin/actions": {
        const body = request.postDataJSON();
        commands.push(body);
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
  return { page, context, setAllowed(value) { allowed = value; }, setModelAllowed(value) { modelAllowed = value; } };
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
  await page.getByRole("button", { name: "Select", exact: true }).click();
  assert.equal(await page.getByRole("textbox", { name: /key/ }).inputValue(), "sample.txt");
  await page.getByRole("button", { name: "Query", exact: true }).click();
  await visible(page.getByText("fixture-result", { exact: true }));
  await screenshot(page, "desktop-plugin-inventory");
  // Revocation has no observation event. Both the inventory and the visible
  // query result must still disappear on their next authorization check.
  desktop.setAllowed(false);
  await visible(page.getByText("No resources on this page.", { exact: true }));
  await visible(page.getByRole("alert").filter({ hasText: "Resource permission revoked" }));
  assert.equal(await page.getByText("fixture-result", { exact: true }).count(), 0);
  desktop.setAllowed(true);
  await visible(page.getByText("sample.txt", { exact: true }));

  // A mutation must not run when the operator cancels confirmation.
  await page.getByLabel("Operation").selectOption("files.publish");
  await page.getByRole("textbox", { name: /key/ }).fill("sample.txt");
  page.once("dialog", dialog => dialog.dismiss());
  await page.getByRole("button", { name: "Run action" }).click();
  assert.equal(commands.filter(command => command.operation === "plugin.admin.invoke").length, 0);
  page.once("dialog", dialog => dialog.accept());
  await page.getByRole("button", { name: "Run action" }).click();
  await visible(page.getByText("fixture-result", { exact: false }).last());
  assert.equal(commands.filter(command => command.operation === "plugin.admin.invoke").length, 1);

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

  await page.getByRole("navigation", { name: "Admin navigation" }).getByRole("link", { name: "Projects", exact: true }).click();
  await visible(page.getByRole("heading", { name: "Projects", level: 1 }));
  await page.locator("#selected-project").selectOption("beta");
  await visible(page.locator(".workspace-picker").getByText("beta", { exact: true }));
  await menuSearch(page, "File metadata");
  await visible(page.getByText("No matching pages.", { exact: true }));
  await page.goto(origin + "/beta/ui/extensions/files/0.1.0/operations");
  assert.equal(new URL(page.url()).pathname, "/beta/ui/extensions/files/0.1.0/operations");
  await visible(page.getByRole("alert").filter({ hasText: "Page unavailable or permission revoked" }));
  await desktop.context.close();

  const mobile = await fixture({ width: 390, height: 844 });
  await mobile.page.goto(origin + "/alpha/ui/plugins");
  await visible(mobile.page.getByRole("heading", { name: "Plugins", level: 1 }));
  const toggle = mobile.page.getByRole("button", { name: "Toggle navigation" });
  await visible(toggle);
  assert.equal(await toggle.getAttribute("aria-expanded"), "false");
  await toggle.click();
  await menuSearch(mobile.page, "file metadata");
  await visible(mobile.page.getByRole("navigation", { name: "Admin navigation" }).getByRole("link", { name: /File metadata/ }));
  await screenshot(mobile.page, "mobile-navigation");
  await mobile.page.getByRole("navigation", { name: "Admin navigation" }).getByRole("link", { name: /File metadata/ }).click();
  await visible(mobile.page.getByRole("heading", { name: "File metadata", level: 1 }));
  assert.equal(await toggle.getAttribute("aria-expanded"), "false");
  assert.ok(await mobile.page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth + 1), "mobile page must not overflow horizontally");
  await screenshot(mobile.page, "mobile-plugin");
  await mobile.context.close();

  assert.deepEqual(runtimeErrors, [], "browser runtime errors");
  console.log("Admin browser smoke passed: desktop, mobile, permissions, project switching and action confirmation.");
} finally {
  await browser.close();
}
