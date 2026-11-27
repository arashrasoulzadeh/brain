// CI-only transport into the network-disabled backend container. This never
// exposes a Docker socket or an arbitrary command/URL to the browser.
import { createServer } from "node:http";
import { spawn } from "node:child_process";
import { createInterface } from "node:readline";

const container = process.argv[2];
if (!/^[a-f0-9]{12,64}$/.test(container ?? "")) throw new Error("expected an isolated CI container ID");
const port = 18082;
const childProgram = `
import { createInterface } from "node:readline";
const lines = createInterface({input:process.stdin,crlfDelay:Infinity});
for await (const line of lines) {
  let request;
  try {
    request = JSON.parse(line);
    if (!request.path.startsWith("/") || request.path.startsWith("//")) throw new Error("invalid path");
    const headers = {};
    for (const name of ["authorization","content-type","accept","x-project-id","x-project-slug"]) if (request.headers[name]) headers[name] = request.headers[name];
    const response = await fetch("http://127.0.0.1:8080"+request.path,{method:request.method,headers,body:request.body ? Buffer.from(request.body,"base64") : undefined,signal:AbortSignal.timeout(25000),redirect:"manual"});
    const body = Buffer.from(await response.arrayBuffer());
    const binary = /^\\/[A-Za-z0-9][A-Za-z0-9_-]*\\/admin\\/files\\/blob$/.test(request.path.split("?")[0]);
    if (body.length > (binary ? 24 : 8)*1024*1024) throw new Error("response too large");
    const outgoing = {"content-type":response.headers.get("content-type") ?? "application/octet-stream"};
    for (const name of ["content-disposition","content-length","cache-control","x-content-type-options"]) { const value = response.headers.get(name); if (value !== null) outgoing[name] = value; }
    console.log(JSON.stringify({id:request.id,status:response.status,headers:outgoing,body:body.toString("base64")}));
  } catch {
    console.log(JSON.stringify({id:request?.id,status:502,headers:{"content-type":"text/plain"},body:Buffer.from("isolated backend unavailable").toString("base64")}));
  }
}
`;
const child = spawn("docker",["exec","-i",container,"node","--input-type=module","-e",childProgram],{stdio:["pipe","pipe","inherit"]});
const pending = new Map();
let sequence = 0;
const lines = createInterface({input:child.stdout,crlfDelay:Infinity});
lines.on("line",line=>{
  let reply;
  try { reply=JSON.parse(line); } catch { return; }
  const item=pending.get(reply.id);
  if (!item) return;
  pending.delete(reply.id);
  clearTimeout(item.timer);
  if (!item.response.destroyed) {
    item.response.writeHead(reply.status,reply.headers);
    item.response.end(Buffer.from(reply.body,"base64"));
  }
});
const server=createServer(async (request,response)=>{
  if (!request.url?.startsWith("/") || request.url.startsWith("//") || pending.size >= 32) {
    response.writeHead(400).end(); return;
  }
  try {
    // The test relay buffers requests; keep the normal cap and permit only
    // the reserved binary endpoint enough room for the integration fixture.
    const binary = /^\/[A-Za-z0-9][A-Za-z0-9_-]*\/admin\/files\/blob$/.test(request.url.split("?")[0]);
    const maximum = (binary ? 24 : 1)*1024*1024;
    const chunks=[];
    let size=0;
    for await (const chunk of request) {
      size+=chunk.length;
      if (size>maximum) { response.writeHead(413).end(); return; }
      chunks.push(chunk);
    }
    const id=++sequence;
    const timer=setTimeout(()=>{
      pending.delete(id);
      if (!response.destroyed) response.writeHead(504).end("isolated backend timed out");
    },60000);
    pending.set(id,{response,timer});
    child.stdin.write(JSON.stringify({id,path:request.url,method:request.method,headers:request.headers,body:Buffer.concat(chunks).toString("base64")})+"\n");
  } catch { if (!response.destroyed) response.writeHead(502).end(); }
});
function stop() {
  for (const {response,timer} of pending.values()) {
    clearTimeout(timer);
    if (!response.destroyed) response.writeHead(502).end("isolated backend stopped");
  }
  pending.clear();
  server.close();
  child.kill();
}
child.on("exit",()=>{stop();process.exitCode=1;});
child.on("error",()=>{stop();process.exitCode=1;});
child.stdin.on("error",stop);
process.on("SIGTERM",stop);
process.on("SIGINT",stop);
server.listen(port,"127.0.0.1",()=>console.log("Isolated backend CI relay listening on loopback"));
