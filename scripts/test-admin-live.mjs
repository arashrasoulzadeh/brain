import assert from "node:assert/strict";
import { mkdir } from "node:fs/promises";
import { createRequire } from "node:module";
import { resolve } from "node:path";
import { setTimeout as delay } from "node:timers/promises";

const require = createRequire(resolve(process.env.BRAIN_BROWSER_DEPS ?? ".", "package.json"));
const { chromium } = require("playwright");
const origin = process.env.BRAIN_ADMIN_ORIGIN ?? "http://127.0.0.1:3100";
const backend = process.env.BRAIN_API_ORIGIN ?? "http://127.0.0.1:18082";
const output = resolve(process.env.BRAIN_BROWSER_OUTPUT ?? ".local/admin-browser");
const username = "ci-operator", secret = "ci-fixture-only", project = "browser-live", other = "browser-other";
const version = "0.1.0";

let ready = false;
for (let attempt = 0; attempt < 30; attempt++) {
  try {
    ready = (await fetch(backend + "/_system/readyz", { signal: AbortSignal.timeout(1000) })).ok;
    if (ready) break;
  } catch {}
  await delay(1000);
}
assert.ok(ready, "packaged backend never became ready");
const login = await fetch(backend + "/_system/admin/session", {method:"POST", headers:{"content-type":"application/json"}, body:JSON.stringify({username,secret})});
assert.ok(login.ok, "fixture operator login failed");
const {session} = await login.json();
const headers = {authorization:"Bearer " + session, "content-type":"application/json"};
async function action(operation, extra = {}) {
  const response = await fetch(backend + "/_system/admin/actions", {method:"POST", headers,
    body:JSON.stringify({operation,project_id:project,plugin_id:"files",version,...extra}), signal:AbortSignal.timeout(20000)});
  assert.ok(response.ok, operation + ": " + await response.clone().text());
  return response.json();
}
function decode(result) {
  assert.equal(result.Success, true);
  return JSON.parse(Buffer.from(result.Payload,"base64").toString("utf8"));
}
async function install(plugin_id, capabilities, project_id = project) {
  const identity = {plugin_id,project_id};
  await action("plugin.install", identity);
  await action("plugin.readiness", {...identity,ready:true});
  for (const capability of capabilities) await action("plugin.grant", {...identity,capability});
  await action("plugin.enable", identity);
}
await action("project.create");
await action("project.create", {project_id:other});
for (const project_id of [project,other]) {
  await install("files", ["files.upload","files.download","files.metadata","files.publish"], project_id);
  for (const key of project_id === project ? ["visible.txt","hidden.txt"] : ["foreign.txt"]) {
    decode(await action("plugin.invoke", {project_id,target_operation:"files.upload",payload:{key,data:"fixture content"}}));
  }
}
// Enroll the project after seeding; only explicit operator policy permits reads.
await install("access-management", ["access.rules.read","access.rules.manage"]);
async function rule(id, operation, resource, allowed = true) {
  return decode(await action("plugin.admin.invoke", {plugin_id:"access-management",page_id:"rules",confirmed:true,
    target_operation:"access.rules.create",payload:{id,fields:{principal:"operator:"+username,plugin:"files",version,operation,capability:operation,resource,allowed}}}));
}
const readRule = await rule("read-visible","files.metadata","visible.txt");
await rule("publish-visible","files.publish","visible.txt");
await rule("deny-hidden","files.metadata","hidden.txt",false);
await mkdir(output,{recursive:true});
const browser = await chromium.launch({ headless: true, executablePath: process.env.BRAIN_BROWSER_EXECUTABLE || undefined });
const context = await browser.newContext({viewport:{width:1440,height:900}});
const page = await context.newPage();
const runtimeErrors = [];
page.on("pageerror",error=>runtimeErrors.push(error.message));
async function visible(locator) { await locator.waitFor({state:"visible",timeout:20000}); }
async function screenshot(name) { await page.screenshot({path:resolve(output,name+".png"),fullPage:true,animations:"disabled"}); }
try {
  await page.goto(origin);
  await visible(page.getByRole("heading",{name:"Welcome back"}));
  await page.getByLabel("Operator username").fill(username);
  await page.getByLabel("Secret",{exact:true}).fill(secret);
  await page.getByRole("button",{name:"Sign in",exact:true}).click();
  await visible(page.getByRole("heading",{name:"Projects",level:1}));
  await page.locator("#selected-project").selectOption(project);
  await visible(page.locator(".workspace-picker").getByText(project,{exact:true}));
  await page.getByRole("navigation",{name:"Admin navigation"}).getByRole("link",{name:"Files",exact:true}).click();
  const inventory = page.locator(".admin-inventory-panel");
  await visible(inventory.getByText("visible.txt",{exact:true}));
  assert.equal(await inventory.getByText("hidden.txt",{exact:true}).count(),0);
  assert.equal(await inventory.getByText("foreign.txt",{exact:true}).count(),0);
  await inventory.getByRole("button",{name:"Select",exact:true}).click();
  assert.equal(await page.getByRole("textbox",{name:/key/}).inputValue(),"visible.txt");
  await page.getByRole("button",{name:"Query",exact:true}).click();
  await visible(page.getByRole("heading",{name:"Size (bytes)"}));
  await screenshot("live-policy-inventory");

  await page.getByLabel("Operation").selectOption("files.publish");
  await page.getByRole("textbox",{name:/key/}).fill("visible.txt");
  page.once("dialog",dialog=>dialog.dismiss());
  await page.getByRole("button",{name:"Run action"}).click();
  let stored = decode(await action("plugin.admin.query",{page_id:"operations",target_operation:"files.metadata",payload:{key:"visible.txt"}}));
  assert.equal(stored.published,false,"cancelled action changed persisted state");
  page.once("dialog",dialog=>dialog.accept());
  await page.getByRole("button",{name:"Run action"}).click();
  await visible(inventory.getByText("Yes",{exact:true}));
  stored = decode(await action("plugin.admin.query",{page_id:"operations",target_operation:"files.metadata",payload:{key:"visible.txt"}}));
  assert.equal(stored.published,true);
  await screenshot("live-published-file");

  // A real policy revocation must remove the resource on the next inventory read.
  decode(await action("plugin.admin.invoke",{plugin_id:"access-management",page_id:"rules",confirmed:true,
    target_operation:"access.rules.revoke",payload:{id:"read-visible",revision:readRule.revision}}));
  await inventory.getByRole("button",{name:"Refresh",exact:true}).click();
  await visible(inventory.getByText("No resources on this page.",{exact:true}));
  assert.equal(await inventory.getByText("visible.txt",{exact:true}).count(),0);
  const denied = await fetch(backend+"/_system/admin/actions",{method:"POST",headers,body:JSON.stringify({operation:"plugin.admin.query",
    project_id:project,plugin_id:"files",version,page_id:"operations",target_operation:"files.metadata",payload:{key:"visible.txt"}})});
  assert.equal(denied.status,403);

  await page.getByRole("navigation",{name:"Admin navigation"}).getByRole("link",{name:"Projects",exact:true}).click();
  await page.locator("#selected-project").selectOption(other);
  await visible(page.locator(".workspace-picker").getByText(other,{exact:true}));
  await page.getByRole("navigation",{name:"Admin navigation"}).getByRole("link",{name:"Files",exact:true}).click();
  await visible(inventory.getByText("foreign.txt",{exact:true}));
  assert.equal(await inventory.getByText("visible.txt",{exact:true}).count(),0);
  await screenshot("live-project-isolation");
  await page.getByRole("button",{name:"Sign out",exact:true}).click();
  await visible(page.getByRole("heading",{name:"Welcome back"}));
  assert.equal(await page.evaluate(()=>sessionStorage.getItem("brain-admin-session")),null);
  assert.deepEqual(runtimeErrors,[]);
  console.log("Live admin browser smoke passed: real login, policy filtering, confirmed persistence, revocation and project isolation.");
} finally {
  await context.close();
  await browser.close();
}
