import assert from "node:assert/strict";
import { mkdir, readFile } from "node:fs/promises";
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
  const body = {operation,project_id:project,plugin_id:"files",version,...extra};
  const namespace = operation === "project.create" || operation === "plugin.readiness" ? "_system" : encodeURIComponent(body.project_id);
  const response = await fetch(backend + "/" + namespace + "/admin/actions", {method:"POST", headers,
    body:JSON.stringify(body), signal:AbortSignal.timeout(20000)});
  assert.ok(response.ok, operation + ": " + await response.clone().text());
  return response.json();
}
function decode(result) {
  assert.equal(result.Success, true);
  return JSON.parse(Buffer.from(result.Payload,"base64").toString("utf8"));
}
async function install(plugin_id, capabilities, project_id = project) {
  const identity = {plugin_id,project_id};
  const catalogResponse = await fetch(backend+"/"+project_id+"/admin/plugins",{headers});
  assert.ok(catalogResponse.ok);
  const catalog = await catalogResponse.json();
  const manifest = catalog.manifests.find(item=>item.ID===plugin_id && item.Version===version);
  assert.ok(manifest);
  for (const capability of capabilities) assert.ok(manifest.Capabilities.includes(capability));
  const command = {...identity,operation:"plugin.install.activate",version,capabilities:manifest.Capabilities};
  const missing = await fetch(backend+"/"+project_id+"/admin/actions",{method:"POST",headers,body:JSON.stringify(command)});
  assert.equal(missing.status,400);
  const stale = await fetch(backend+"/"+project_id+"/admin/actions",{method:"POST",headers,body:JSON.stringify({...command,confirmed:true,capabilities:["unapproved"]})});
  assert.equal(stale.status,403);
  await action("plugin.install.activate",{...identity,confirmed:true,capabilities:manifest.Capabilities});
  const refreshed = await fetch(backend+"/"+project_id+"/admin/plugins",{headers});
  const status = (await refreshed.json()).installations.find(item=>item.PluginID===plugin_id && item.Version===version);
  assert.ok(status.Enabled && status.Ready && status.WorkerReady);
  assert.deepEqual([...status.Grants].sort(),[...manifest.Capabilities].sort());
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
  await page.goto(origin + "/_system/ui");
  await visible(page.getByRole("heading",{name:"Welcome back"}));
  await page.getByLabel("Operator username").fill(username);
  await page.getByLabel("Secret",{exact:true}).fill(secret);
  await page.getByRole("button",{name:"Sign in",exact:true}).click();
  await visible(page.getByRole("heading",{name:"Projects",level:1}));
  await page.locator("#selected-project").selectOption(project);
  await visible(page.locator(".workspace-picker").getByText(project,{exact:true}));
  const filesLink = page.getByRole("navigation",{name:"Admin navigation"}).getByRole("link",{name:"Files",exact:true});
  assert.equal(await filesLink.getAttribute("href"), "/" + project + "/ui/files");
  await Promise.all([
    page.waitForURL(origin + "/" + project + "/ui/files", {timeout:20000}),
    filesLink.click(),
  ]);
  assert.equal(new URL(page.url()).pathname, "/" + project + "/ui/files");
  const inventory = page.locator(".admin-inventory-panel");
  await visible(inventory.locator("td:first-child").getByText("visible.txt",{exact:true}));
  assert.equal(await inventory.locator("td:first-child").getByText("hidden.txt",{exact:true}).count(),0);
  assert.equal(await inventory.locator("td:first-child").getByText("foreign.txt",{exact:true}).count(),0);
  await inventory.getByRole("button",{name:"Select",exact:true}).click();
  await visible(page.getByRole("dialog",{name:"File metadata"}));
  assert.equal(await page.getByRole("textbox",{name:/key/}).inputValue(),"visible.txt");
  await page.getByRole("button",{name:"Query",exact:true}).click();
  await visible(page.getByRole("heading",{name:"Size (bytes)"}));
  await screenshot("live-policy-inventory");

  await page.getByLabel("Operation").selectOption("files.publish");
  await page.getByRole("textbox",{name:/key/}).fill("visible.txt");
  await page.getByRole("button",{name:"Run action"}).click();
  await page.getByRole("button",{name:"Cancel confirmation"}).click();
  let stored = decode(await action("plugin.admin.query",{page_id:"operations",target_operation:"files.metadata",payload:{key:"visible.txt"}}));
  assert.equal(stored.published,false,"cancelled action changed persisted state");
  await page.getByRole("button",{name:"Run action"}).click();
  await page.getByRole("button",{name:"Confirm action"}).click();
  await visible(page.getByRole("status").filter({hasText:"Action completed"}));
  await page.getByRole("button",{name:"Close dialog"}).click();
  await visible(inventory.getByText("Yes",{exact:true}));
  stored = decode(await action("plugin.admin.query",{page_id:"operations",target_operation:"files.metadata",payload:{key:"visible.txt"}}));
  assert.equal(stored.published,true);
  await screenshot("live-published-file");

  // A real policy revocation must remove the resource on the next inventory read.
  decode(await action("plugin.admin.invoke",{plugin_id:"access-management",page_id:"rules",confirmed:true,
    target_operation:"access.rules.revoke",payload:{id:"read-visible",revision:readRule.revision}}));
  await inventory.getByRole("button",{name:"Refresh",exact:true}).click();
  await visible(inventory.getByText("No resources on this page.",{exact:true}));
  assert.equal(await inventory.locator("td:first-child").getByText("visible.txt",{exact:true}).count(),0);
  const denied = await fetch(backend+"/"+encodeURIComponent(project)+"/admin/actions",{method:"POST",headers,body:JSON.stringify({operation:"plugin.admin.query",
    project_id:project,plugin_id:"files",version,page_id:"operations",target_operation:"files.metadata",payload:{key:"visible.txt"}})});
  assert.equal(denied.status,403);

  await page.getByRole("navigation",{name:"Admin navigation"}).getByRole("link",{name:"Projects",exact:true}).click();
  await page.locator("#selected-project").selectOption(other);
  await visible(page.locator(".workspace-picker").getByText(other,{exact:true}));
  assert.equal(await filesLink.getAttribute("href"), "/" + other + "/ui/files");
  await Promise.all([
    page.waitForURL(origin + "/" + other + "/ui/files", {timeout:20000}),
    filesLink.click(),
  ]);
  await visible(inventory.locator("td:first-child").getByText("foreign.txt",{exact:true}));
  assert.equal(new URL(page.url()).pathname, "/" + other + "/ui/files");
  assert.equal(await inventory.locator("td:first-child").getByText("visible.txt",{exact:true}).count(),0);
  await screenshot("live-project-isolation");

  // Change the Files limit through the exact project installation settings UI.
  const manage = page.getByRole("link",{name:"Manage files in "+other});
  assert.equal(await manage.getAttribute("href"),"/"+other+"/ui/plugins/files%400.1.0");
  await manage.click();
  await visible(page.getByRole("heading",{name:"Project settings"}));
  await page.getByRole("button",{name:"Edit settings"}).click();
  await visible(page.getByRole("dialog",{name:"Edit project settings"}));
  await page.getByLabel("Maximum file size (MB)").fill("0.001024");
  await page.getByRole("button",{name:"Save configuration"}).click();
  await page.getByRole("button",{name:"Confirm settings"}).click();
  await visible(page.getByRole("status").filter({hasText:"Configuration updated"}));
  await page.getByRole("button",{name:"Close dialog"}).click();
  await visible(page.getByText("0.001024 MB",{exact:true}));
  // The setting must not leak into another project or expose stored secrets.
  for (const [id, limit] of [[project,100000000],[other,1024]]) {
    const response = await fetch(backend+"/"+id+"/admin/plugins",{headers});
    assert.ok(response.ok);
    const catalog = await response.json();
    assert.equal(catalog.installations.find(item=>item.PluginID==="files").MaxUploadBytes,limit);
  }
  const rejected = await fetch(backend+"/"+other+"/admin/actions",{method:"POST",headers,body:JSON.stringify({operation:"plugin.admin.invoke",project_id:other,plugin_id:"files",version,page_id:"operations",target_operation:"files.upload",confirmed:true,payload:{key:"too-big",data:"x".repeat(1025)}})});
  assert.equal(rejected.ok,false);
  await filesLink.click();
  await visible(inventory.locator("td:first-child").getByText("foreign.txt",{exact:true}));

  // File picker sends the existing authorized string contract, without losing binary bytes.
  const bytes = Buffer.from([0, 255, 128, 13, 10, 65, 0]);
  const filename = "browser-upload.bin";
  const key = "custom-key-without-extension";
  await page.getByRole("button",{name:"Upload file",exact:true}).click();
  await page.getByLabel("Local file").setInputFiles({name:"too-large.bin",mimeType:"application/octet-stream",buffer:Buffer.alloc(1025)});
  await visible(page.getByRole("alert").filter({hasText:"0.001024 MB"}));
  await page.getByLabel("Local file").setInputFiles({name:filename,mimeType:"application/octet-stream",buffer:bytes});
  await visible(page.getByRole("status").filter({hasText:"ready to upload"}));
  assert.equal(await page.getByRole("textbox",{name:/key/}).inputValue(),filename);
  await page.getByRole("textbox",{name:/key/}).fill(key);
  assert.equal(await page.getByRole("textbox",{name:"filename"}).inputValue(),filename);
  await page.getByRole("button",{name:"Run action"}).click();
  await page.getByRole("button",{name:"Cancel confirmation"}).click();
  assert.equal(await inventory.locator("td:first-child").getByText(key,{exact:true}).count(),0);
  await page.getByRole("button",{name:"Run action"}).click();
  await page.getByRole("button",{name:"Confirm action"}).click();
  await visible(page.getByRole("status").filter({hasText:"File uploaded successfully"}));
  await page.getByRole("button",{name:"Close dialog"}).click();
  await visible(inventory.locator("td:first-child").getByText(key,{exact:true}));
  const uploaded = decode(await action("plugin.admin.query",{project_id:other,page_id:"operations",target_operation:"files.download",payload:{key}}));
  assert.equal(uploaded.data,"data:application/octet-stream;base64," + bytes.toString("base64"));
  assert.equal(uploaded.metadata.filename,filename);
  assert.equal(uploaded.metadata.extension,"bin");
  assert.equal(uploaded.metadata.size_bytes,bytes.length);
  assert.equal(uploaded.metadata.declared_content_type,"application/octet-stream");
  assert.match(uploaded.metadata.sha256,/^[a-f0-9]{64}$/);
  assert.ok(uploaded.metadata.created_at && uploaded.metadata.updated_at);
  const meta = decode(await action("plugin.admin.query",{project_id:other,page_id:"operations",target_operation:"files.metadata",payload:{key}}));
  assert.deepEqual(meta.metadata,uploaded.metadata);
  assert.equal(meta.size,bytes.length);
  assert.equal("data" in meta,false);
  await visible(inventory.getByText(filename,{exact:true}));
  await page.getByRole("button",{name:"Download file",exact:true}).click();
  await page.getByRole("textbox",{name:/key/}).fill(key);
  await page.getByRole("button",{name:"Query",exact:true}).click();
  await visible(page.getByRole("button",{name:"Save original file",exact:true}));
  const downloadPromise = page.waitForEvent("download");
  await page.getByRole("button",{name:"Save original file",exact:true}).click();
  const download = await downloadPromise;
  assert.equal(download.suggestedFilename(),filename);
  assert.deepEqual(await readFile(await download.path()),bytes);
  await screenshot("live-file-upload-download");
  await page.getByRole("button",{name:"Close dialog"}).click();
  // A larger upload crosses the JSON and worker payload ceilings. The raw
  // binary transport preserves content and metadata without increasing either.
  const configure = async value => action("plugin.configure",{project_id:other,values:{maxUploadBytes:value}});
  const badLimit = await fetch(backend+"/"+other+"/admin/actions",{method:"POST",headers,body:JSON.stringify({operation:"plugin.configure",project_id:other,plugin_id:"files",version,values:{maxUploadBytes:1000000001}})});
  assert.equal(badLimit.ok,false);
  await configure(20000000);
  await page.reload();
  await visible(inventory.locator("td:first-child").getByText(key,{exact:true}));
  const bigKey = "binary-large";
  const large = Buffer.alloc(12 * 1024 * 1024 + 17,0xab);
  large[0]=0; large[large.length-1]=42;
  const blobURL = (id, object, more={}) => backend+"/"+id+"/admin/files/blob?"+new URLSearchParams({version,key:object,...more});
  const binaryHeaders = {...headers,"content-type":"application/octet-stream"};
  const unauth = await fetch(blobURL(other,bigKey));
  assert.equal(unauth.status,401);
  const oversized = await fetch(blobURL(other,bigKey,{confirmed:"true",filename:"large.bin"}),{method:"POST",headers:binaryHeaders,body:Buffer.alloc(20000001)});
  assert.equal(oversized.status,413);
  const missingConfirm = await fetch(blobURL(other,bigKey),{method:"POST",headers:binaryHeaders,body:Buffer.from("confirmation required")});
  assert.equal(missingConfirm.status,400,await missingConfirm.text());
  const blocked = await fetch(blobURL(project,bigKey,{confirmed:"true",filename:"large.bin"}),{method:"POST",headers:binaryHeaders,body:Buffer.from("denied")});
  assert.equal(blocked.status,403,await blocked.text());
  await page.getByRole("button",{name:"Upload file",exact:true}).click();
  await page.getByLabel("Local file").setInputFiles({name:"large.bin",mimeType:"application/octet-stream",buffer:large});
  await visible(page.getByRole("status").filter({hasText:"ready to upload"}));
  await page.getByRole("textbox",{name:/key/}).fill(bigKey);
  await page.getByRole("button",{name:"Run action"}).click();
  await page.getByRole("button",{name:"Confirm action"}).click();
  try {
    await visible(page.getByRole("status").filter({hasText:"File uploaded successfully"}));
  } catch (reason) {
    console.error("Binary upload dialog:", await page.getByRole("dialog").innerText().catch(()=>"not open"));
    await screenshot("binary-upload-failure");
    throw reason;
  }
  await page.getByRole("button",{name:"Close dialog"}).click();
  await visible(inventory.locator("td:first-child").getByText(bigKey,{exact:true}));
  const binary = decode(await action("plugin.admin.query",{project_id:other,page_id:"operations",target_operation:"files.download",payload:{key:bigKey}}));
  assert.equal(binary.streamed,true);
  assert.equal(binary.metadata.encoding,"binary");
  assert.equal(binary.metadata.size_bytes,large.length);
  assert.equal(binary.metadata.filename,"large.bin");
  assert.equal(binary.data,undefined);
  const foreign = await fetch(blobURL(project,bigKey),{headers});
  assert.equal(foreign.status,403);
  await page.getByRole("button",{name:"Download file",exact:true}).click();
  await page.getByRole("textbox",{name:/key/}).fill(bigKey);
  await page.getByRole("button",{name:"Query",exact:true}).click();
  await visible(page.getByRole("button",{name:"Save original file",exact:true}));
  const largeDownload = page.waitForEvent("download");
  await page.getByRole("button",{name:"Save original file",exact:true}).click();
  const saved = await largeDownload;
  assert.equal(saved.suggestedFilename(),"large.bin");
  assert.deepEqual(await readFile(await saved.path()),large);
  await page.getByRole("button",{name:"Close dialog"}).click();
  decode(await action("plugin.admin.invoke",{project_id:other,page_id:"operations",target_operation:"files.publish",payload:{key:bigKey},confirmed:true}));
  const published = await fetch(blobURL(other,bigKey),{headers});
  assert.equal(published.status,200);
  assert.equal(Number(published.headers.get("content-length")),large.length);
  assert.deepEqual(Buffer.from(await published.arrayBuffer()),large);
  await page.getByRole("button",{name:"Sign out",exact:true}).click();
  await visible(page.getByRole("heading",{name:"Welcome back"}));
  assert.equal(await page.evaluate(()=>sessionStorage.getItem("brain-admin-session")),null);
  assert.deepEqual(runtimeErrors,[]);
  console.log("Live admin browser smoke passed: real login, policy filtering, confirmed persistence, revocation and project isolation.");
} finally {
  await context.close();
  await browser.close();
}
