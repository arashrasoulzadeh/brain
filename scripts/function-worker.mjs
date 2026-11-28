import vm from "node:vm";
import readline from "node:readline";
import { readFile } from "node:fs/promises";

const descriptor = JSON.parse(await readFile(process.argv[2], "utf8"));
if (descriptor.pluginId !== "functions" || !descriptor.projectId || !descriptor.version || descriptor.routes?.length !== 1 || !["functions.call", "functions.invoke"].includes(descriptor.routes[0].operation)) throw new Error("invalid function descriptor");
const source = await readFile(descriptor.entrypoint, "utf8");
const helperSource = await readFile(new URL("../plugins/sdk-typescript/src/helpers.mjs", import.meta.url), "utf8");
const clock = () => Number(process.hrtime.bigint()) / 1e6;
const denyImport = () => { throw new Error("imports are not permitted"); };
process.stdout.write('{"ready":true}\n');
const lines = readline.createInterface({ input: process.stdin, crlfDelay: Infinity });
const iterator = lines[Symbol.asyncIterator]();
const first = await iterator.next();
if (!first.done) {
  const line = first.value;
  let request;
  try {
    request = JSON.parse(line);
    const p = request.params;
    if (Buffer.byteLength(line) > 1048576 || p.protocol !== 1 || p.projectId !== descriptor.projectId || p.project?.id !== descriptor.projectId || p.project.active !== true || p.pluginId !== "functions" || p.version !== descriptor.version || p.operation !== descriptor.routes[0].operation || p.hostContext?.projectVerified !== true || !p.hostContext.invocationId) throw new Error("invalid function invocation");
    const sandbox = vm.createContext(Object.create(null), { codeGeneration: { strings: false, wasm: false } });
    const helpers = new vm.SourceTextModule(helperSource, { context: sandbox });
    await helpers.link(denyImport);
    await helpers.evaluate({ timeout: 450 });
    sandbox.__helper = helpers.namespace.callHelper;
    const run = (code) => vm.runInContext(code, sandbox, { timeout: 450 });
    const json = (value) => "JSON.parse(" + JSON.stringify(JSON.stringify(value)) + ")";
    sandbox.__input = run(json(p.payload ?? null));
    sandbox.__auth = run("(()=>{const a=" + json(p.hostContext.auth) + ";const permissions=Object.freeze(a.permissions.map(p=>Object.freeze(p)));return Object.freeze({actor:Object.freeze(a.actor),caller:Object.freeze(a.caller),permissions,can:(capability,resource='')=>permissions.some(p=>p.capability===capability&&p.resource===resource)})})()");
    // Keep the host callback in a private VM-realm closure. Only primitive
    // JSON strings cross the realm boundary; no outer errors or objects do.
    const draft = p.operation === "functions.invoke" && p.hostContext.auth.actor.kind === "operator";
    let tail = Promise.resolve();
    let callID = 0;
    const hostCall = raw => {
      const job = tail.then(async () => {
        if (!draft) return '{"error":"Plugin calls require an operator sandbox test"}';
        if (typeof raw !== "string" || Buffer.byteLength(raw) > 262144 || callID >= 16) return '{"error":"Plugin call limit exceeded"}';
        const { operation, input } = JSON.parse(raw);
        if (typeof operation !== "string" || !operation || operation.length > 128) return '{"error":"Invalid plugin operation"}';
        const id = ++callID;
        process.stdout.write(JSON.stringify({hostCall:{id,operation,input:input??null}}) + "\n");
        const response = await iterator.next();
        if (response.done || Buffer.byteLength(response.value) > 1048576) throw new Error("host bridge closed");
        const envelope = JSON.parse(response.value).hostResponse;
        if (!envelope || envelope.id !== id) throw new Error("invalid host response");
        return JSON.stringify(envelope.error ? {error:envelope.error} : {result:envelope.output});
      }).catch(() => '{"error":"Plugin bridge unavailable"}');
      tail = job.then(() => {});
      return job;
    };
    const makeBridge = run("(hostCall)=>{const stringify=JSON.stringify,parse=JSON.parse,Failure=Error;return async (operation,input)=>{if(typeof operation!=='string')throw new Failure('Plugin operation must be a string');const result=parse(await hostCall(stringify({operation,input})));if(result.error)throw new Failure(result.error);return result.result;}}");
    sandbox.__execute = makeBridge(hostCall);
    sandbox.__context = run("Object.freeze({auth:__auth,project:Object.freeze(" + json({id:descriptor.projectId,active:true}) + "),plugin:Object.freeze({id:'functions'}),execution:Object.freeze(" + json({operation:p.operation,id:p.hostContext.invocationId,invocationId:p.hostContext.invocationId}) + "),helpers:Object.freeze({call:__helper}),execute:__execute,plugins:Object.freeze({call:__execute})})");
    const compileStart = clock();
    const module = new vm.SourceTextModule(source, { context: sandbox });
    await module.link(denyImport);
    const compileMs = clock() - compileStart;
    const evaluateStart = clock();
    await module.evaluate({ timeout: 450 });
    const evaluateMs = clock() - evaluateStart;
    if (typeof module.namespace.default !== "function") throw new Error("default export must be a function");
    sandbox.__handler = module.namespace.default;
    const invokeStart = clock();
    let timer;
    try {
      sandbox.__result = await Promise.race([Promise.resolve(run("__handler(__input,__context)")), new Promise((_, reject) => {timer = setTimeout(() => reject(new Error("function execution timed out")), draft ? 10000 : 400);})]);
      await tail;
      const result = JSON.parse(run("JSON.stringify(__result??null)"));
      const payload = {result, timings:{compile_ms:compileMs,evaluate_ms:evaluateMs,invoke_ms:clock()-invokeStart}};
      const encoded = JSON.stringify({jsonrpc:"2.0",id:request.id,result:{protocol:1,success:true,payload}});
      if (Buffer.byteLength(encoded) > 1048576) throw new Error("result too large");
      process.stdout.write(encoded + "\n");
    } finally { clearTimeout(timer); }
  } catch {
    process.stdout.write(JSON.stringify({jsonrpc:"2.0",id:request?.id,result:{protocol:1,success:false,errorCode:"function execution failed"}}) + "\n");
  }
  // Keep stdout alive until the host closes the sandbox, avoiding a race
  // with exec.Cmd.Wait closing its pipe before the final result is read.
  await iterator.next();
}
lines.close();
