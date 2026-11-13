import { APPROVED_SDK_IMPORTS, APPROVED_WASI_IMPORTS, createApprovalStore, createSharedPackageStore, createFunctionDraftStore, createJobLeaseStore, createInvocationEffects, capDiagnostics, commitWithGrant, compilerEnvironment, createInvocationRequest, diagnoseSource, DisposableCompilerWorkspace, ENTRY_PACKAGE_FIXTURE, evaluateCondition, executeFixture, FunctionPublicationStore, functionRequest, invokeDirectHttps, InvocationCommitStore, loadDeclaredInputs, parseInvocationResult, PROJECT_HELPER_FIXTURE, RecordChangeTriggerStore, redactRunStatus, rejectSandboxEscapeAttempt, requireSDKImports, retryInvocation, runCappedRetries, runIdempotentWorkflow, runWithTimeout, ScheduleStore, validateBuildOverrides, validateCompilerEnvironment, validateCompilerLimits, validateFunctionLimits, validateInvocationResult, validateLinearMemory, validateSandboxCapabilities, validateSDKImports, validateSourceProject, validateStaleInputVersions, validateStartupDuration, validateWasiImports, validateWorkflowGraph, validateWriteIntents, wasmBuildSettings } from "../../plugins/functions/index.js";
validateSDKImports(APPROVED_SDK_IMPORTS);
const request = functionRequest("dev", "functions", "0.1.0", "functions.invoke", { name: "hello" });
if (request.params.projectId !== "dev" || request.params.operation !== "functions.invoke") throw new Error("functions request contract failed");
try { validateSDKImports(["brain/sdk", "os"]); throw new Error("unsupported SDK import accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "unsupported SDK import") throw error; }
if (!ENTRY_PACKAGE_FIXTURE.includes("package main")) throw new Error("entry package fixture contract failed");
if (!PROJECT_HELPER_FIXTURE.includes("package helper")) throw new Error("project helper fixture contract failed");
const wasm = wasmBuildSettings();
if (wasm.goos !== "wasip1" || wasm.goarch !== "wasm" || wasm.cgoEnabled !== "0") throw new Error("wasm build settings failed");
requireSDKImports(["brain/sdk"], ["brain/sdk"]);
try { requireSDKImports([], ["brain/sdk"]); throw new Error("missing SDK import accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "required SDK import missing") throw error; }
validateSourceProject("dev", "dev");
try { validateSourceProject("dev", "prod"); throw new Error("cross-project source accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "cross-project source access denied") throw error; }
validateBuildOverrides({});
try { validateBuildOverrides({ GOWORK: "off" }); throw new Error("workspace override accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "build override denied") throw error; }
try { validateBuildOverrides({ toolchainPath: "/tmp/go" }); throw new Error("toolchain override accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "build override denied") throw error; }
validateFunctionLimits(10, 20, 10, 20);
try { validateFunctionLimits(11, 20, 10, 20); throw new Error("oversized function source accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "function size limit exceeded") throw error; }
const workspace = new DisposableCompilerWorkspace();
if (!workspace.path) throw new Error("compiler workspace was not created");
workspace.dispose();
const environment = compilerEnvironment(); validateCompilerEnvironment(environment);
try { validateCompilerEnvironment({ API_TOKEN: "secret" }); throw new Error("compiler secret accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "compiler credential or network environment denied") throw error; }
validateCompilerLimits({ cpuMs: 1000, memoryBytes: 64 * 1024 * 1024, processes: 4, wallMs: 5000 });
try { validateCompilerLimits({ cpuMs: 0, memoryBytes: 1, processes: 1, wallMs: 1 }); throw new Error("invalid compiler limits accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "invalid compiler limits") throw error; }
const invocation = createInvocationRequest({ projectId: "dev", functionId: "hello", version: "1.0.0", inputs: { name: "alice" }, idempotencyKey: "invoke-1" });
try { createInvocationRequest({ projectId: "bad/project", functionId: "hello", version: "1.0.0", inputs: {}, idempotencyKey: "invoke-2" }); throw new Error("invalid invocation project accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "invalid project id") throw error; }
if (invocation.projectId !== "dev" || invocation.inputs.name !== "alice") throw new Error("typed invocation request failed");
if (!validateInvocationResult({ success: true, output: { ok: true }, diagnostics: [] }).success) throw new Error("typed invocation result failed");
try { validateInvocationResult({ success: true, error: "bad", diagnostics: [] }); throw new Error("invalid invocation result accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "invalid function invocation result") throw error; }
if (!parseInvocationResult({ success: false, error: "failed", diagnostics: [] }).error) throw new Error("parsed invocation result failed");
try { parseInvocationResult(null); throw new Error("malformed function output accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "malformed function output") throw error; }
try { parseInvocationResult({ success: true, diagnostics: [], extra: 1 }); throw new Error("extra invocation result accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "extra function result rejected") throw error; }
if (capDiagnostics(["a", "b", "c"], 2).length !== 2) throw new Error("diagnostic cap failed");
try { capDiagnostics(["x".repeat(5)], 2, 4); throw new Error("oversized diagnostic accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "diagnostic limit exceeded") throw error; }
validateWasiImports(APPROVED_WASI_IMPORTS);
try { validateWasiImports(["wasi_snapshot_preview1", "sock"]); throw new Error("unexpected WASI import accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "unexpected WASI import") throw error; }
validateSandboxCapabilities([]);
try { validateSandboxCapabilities(["filesystem"]); throw new Error("filesystem capability accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "sandbox capability denied") throw error; }
if (await runWithTimeout(Promise.resolve("ok"), 20) !== "ok") throw new Error("function timeout runner failed");
try { await runWithTimeout(new Promise(() => {}), 5); throw new Error("infinite invocation was not timed out"); } catch (error) { if (!(error instanceof Error) || error.message !== "function invocation timed out") throw error; }
validateLinearMemory(1024, 2048);
try { validateLinearMemory(4096, 2048); throw new Error("linear memory limit ignored"); } catch (error) { if (!(error instanceof Error) || error.message !== "linear memory limit exceeded") throw error; }
validateStartupDuration(10, 20);
try { validateStartupDuration(21, 20); throw new Error("startup timeout ignored"); } catch (error) { if (!(error instanceof Error) || error.message !== "startup timeout exceeded") throw error; }
rejectSandboxEscapeAttempt("brain/input");
try { rejectSandboxEscapeAttempt("../secret"); throw new Error("sandbox escape accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "sandbox escape attempt rejected") throw error; }
const effects = createInvocationEffects(); const failed = effects.begin(); failed.write({ id: "failed" }); failed.commit(false);
if (effects.list().length !== 0) throw new Error("failed invocation wrote effects");
const succeeded = effects.begin(); succeeded.write({ id: "ok" }); succeeded.commit(true);
if (effects.list().length !== 1) throw new Error("successful invocation effects missing");
const drafts = createFunctionDraftStore(); drafts.save("dev", "hello", "package main", 1000);
if (drafts.get("dev", "hello")?.source !== "package main" || drafts.get("prod", "hello")) throw new Error("function draft isolation failed");
const packages = createSharedPackageStore(); packages.save("dev", "helper", "package helper");
if (packages.load("dev", "helper") !== "package helper") throw new Error("shared package load failed");
try { packages.load("prod", "helper"); throw new Error("cross-project package loaded"); } catch (error) { if (!(error instanceof Error) || error.message !== "shared package not found") throw error; }
if (diagnoseSource("package main\nimport \"os\"").length !== 1 || diagnoseSource("package main\nimport \"os\"")[0].line !== 2) throw new Error("source diagnostics failed");
if (!executeFixture(() => ({ success: true, output: "ok", diagnostics: [] })).success) throw new Error("fixture execution failed");
const fixtureFailure = executeFixture(() => { throw new Error("fixture boom"); });
if (fixtureFailure.success || fixtureFailure.error !== "fixture boom") throw new Error("fixture failure handling failed");
const publicationDrafts = createFunctionDraftStore();
const publicationDraft = publicationDrafts.save("dev", "hello", "package main", 1000);
const publications = new FunctionPublicationStore();
const publication = publications.publish(publicationDraft, "sha256:abc", "go1.23.0", 2000);
if (publication.version !== 1 || publication.artifactDigest !== "sha256:abc" || publication.toolchain !== "go1.23.0" || publications.get("dev", "hello", 1)?.source !== "package main") throw new Error("function publication failed");
const secondPublication = publications.publish({ ...publicationDraft, source: "package main\n// v2" }, "sha256:def", "go1.23.0", 3000);
if (publications.active("dev", "hello")?.version !== secondPublication.version) throw new Error("latest publication was not active");
if (publications.rollback("dev", "hello", 1).version !== 1 || publications.active("dev", "hello")?.artifactDigest !== "sha256:abc") throw new Error("function rollback failed");
try { publications.publish({ ...publicationDraft, source: "package main\n// broken" }, "bad-digest", "go1.23.0", 4000); throw new Error("invalid draft was published"); } catch (error) { if (!(error instanceof Error) || error.message !== "invalid function publication") throw error; }
if (publications.active("dev", "hello")?.version !== 1) throw new Error("failed draft changed active version");
const httpsResponse = await invokeDirectHttps("https://functions.example.test/invoke", { projectId: "dev", functionId: "hello", version: "1", inputs: {}, idempotencyKey: "request-1" }, async (input, init) => {
  if (input !== "https://functions.example.test/invoke" || init?.method !== "POST") throw new Error("HTTPS invocation request malformed");
  return new Response(JSON.stringify({ success: true, output: { ok: true }, diagnostics: [] }), { status: 200 });
});
if (!httpsResponse.success) throw new Error("direct HTTPS invocation failed");
try { await invokeDirectHttps("http://functions.example.test/invoke", {}, async () => new Response()); throw new Error("HTTP invocation accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "HTTPS endpoint required") throw error; }
for (const endpoint of ["https://127.0.0.1/invoke", "https://169.254.169.254/latest/meta-data", "https://localhost/invoke", "https://user:pass@functions.example.test/invoke"]) {
  try { await invokeDirectHttps(endpoint, {}, async () => new Response()); throw new Error("unsafe HTTPS endpoint accepted"); } catch (error) { if (!(error instanceof Error) || !["HTTPS endpoint target denied", "HTTPS endpoint required"].includes(error.message)) throw error; }
}
const declaredInputs = loadDeclaredInputs({ user: "ada", secret: "must-not-load" }, ["user"]);
if (JSON.stringify(declaredInputs) !== JSON.stringify({ user: "ada" })) throw new Error("undeclared input was loaded");
try { loadDeclaredInputs({ user: "ada" }, ["missing"]); throw new Error("missing declared input accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "declared input missing") throw error; }
const writes = validateWriteIntents([{ operation: "set", path: "profile/name", value: "ada" }, { operation: "delete", path: "profile/old" }], ["profile/name", "profile/old"]);
if (writes.length !== 2) throw new Error("write intents were not validated");
try { validateWriteIntents([{ operation: "set", path: "private/token", value: "secret" }], ["profile/name"]); throw new Error("undeclared write accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "write intent not declared") throw error; }
try { validateWriteIntents([{ operation: "set", path: "../escape", value: true }], ["../escape"]); throw new Error("escaping write accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "invalid write intent") throw error; }
validateStaleInputVersions({ profile: 3, inventory: 8 }, { profile: 3, inventory: 8 });
try { validateStaleInputVersions({ profile: 2 }, { profile: 3 }); throw new Error("stale input accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "stale input version") throw error; }
const commits = new InvocationCommitStore();
const firstCommit = commits.commit("request-1", [{ path: "profile/name", value: "ada" }], { ok: true });
const retryCommit = commits.commit("request-1", [{ path: "profile/name", value: "changed" }], { ok: true });
if (JSON.stringify(firstCommit) !== JSON.stringify(retryCommit) || commits.effects().length !== 1) throw new Error("idempotent effect commit failed");
try { commits.commit("request-2", [{ path: "secret", value: "x" }], undefined, false); throw new Error("failed commit accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "invocation commit failed") throw error; }
if (commits.effects().length !== 1 || commits.result("request-2") !== undefined) throw new Error("failed commit mutated state");
let executions = 0;
const firstResponse = await retryInvocation(commits, "lost-response", () => { executions++; return { ok: true }; }, [{ path: "profile/name", value: "ada" }]);
const recoveredResponse = await retryInvocation(commits, "lost-response", () => { executions++; return { ok: false }; }, [{ path: "profile/name", value: "changed" }]);
if (JSON.stringify(firstResponse) !== JSON.stringify(recoveredResponse) || executions !== 1) throw new Error("lost-response retry re-executed function");
const schedules = new ScheduleStore();
try { schedules.add("bad/project", "hello", "* * * * *", 1000); throw new Error("invalid schedule project accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "invalid project id") throw error; }
schedules.add("dev", "hello", "0 * * * *", 1000);
if (schedules.due(1000)[0]?.functionId !== "hello") throw new Error("scheduled function not due");
try { schedules.add("", "hello", "0 * * * *", 1000); throw new Error("invalid schedule accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "invalid project id") throw error; }
const claimed = schedules.claim("dev", "hello", "worker-a", 1000, 5000);
if (!claimed || schedules.claim("dev", "hello", "worker-b", 1000, 5000)) throw new Error("schedule claimed concurrently");
if (!schedules.recordOccurrence("dev", "hello", "2026-11-13T10:00:00.000Z") || schedules.recordOccurrence("dev", "hello", "2026-11-13T10:00:00.000Z")) throw new Error("duplicate UTC occurrence accepted");
const missed = new ScheduleStore();
missed.add("dev", "sync", "0 * * * *", 1000);
missed.add("dev", "sync", "0 * * * *", 2000);
if (missed.enqueueLatestMissed("dev", "sync", 3000)?.dueAt !== 2000 || missed.enqueueLatestMissed("dev", "sync", 3000) !== undefined) throw new Error("latest missed occurrence was not deduplicated");
const pinned = missed.enqueueVersion("dev", "sync", 3, 4000);
if (pinned.functionVersion !== 3 || pinned.functionId !== "sync") throw new Error("function version was not pinned at enqueue");
const triggers = new RecordChangeTriggerStore();
triggers.add("dev", "profiles", "profile.changed", { functionId: "sync", version: 3 });
if (triggers.match("dev", "profiles", "profile.changed").length !== 1 || triggers.match("prod", "profiles", "profile.changed").length !== 0) throw new Error("record change trigger isolation failed");
if (triggers.dispatch("dev", "profiles", "profile.changed", false).length !== 0 || triggers.dispatch("dev", "profiles", "profile.changed", true).length !== 1) throw new Error("uncommitted change triggered work");
if (triggers.dispatch("dev", "profiles", "profile.changed", true, ["sync"]).length !== 0 || triggers.dispatch("dev", "profiles", "profile.changed", true, ["other"]).length !== 1) throw new Error("recursive worker write was not suppressed");
const leases = createJobLeaseStore();
if (!leases.acquire("job-1", "worker-a", 1000, 500) || leases.acquire("job-1", "worker-b", 1100, 500) || !leases.renew("job-1", "worker-a", 1200, 500) || leases.acquire("job-1", "worker-b", 1500, 500)) throw new Error("short job lease failed");
if (!leases.acquire("job-1", "worker-b", 1800, 500)) throw new Error("expired job lease was not reclaimable");
let grantActive = true;
if (commitWithGrant(() => grantActive, () => "committed", "functions.invoke") !== "committed") throw new Error("authorized job commit failed");
grantActive = false;
try { commitWithGrant(() => grantActive, () => "unsafe", "functions.invoke"); throw new Error("revoked grant committed"); } catch (error) { if (!(error instanceof Error) || error.message !== "grant revoked before commit") throw error; }
let attempts = 0;
try { await runCappedRetries(async () => { attempts++; throw new Error("temporary failure"); }, 2); throw new Error("retry cap ignored"); } catch (error) { if (!(error instanceof Error) || error.message !== "temporary failure" || attempts !== 2) throw error; }
validateWorkflowGraph([{ id: "start", next: ["finish"] }, { id: "finish", next: [] }]);
try { validateWorkflowGraph([{ id: "a", next: ["b"] }, { id: "b", next: ["a"] }]); throw new Error("workflow cycle accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "workflow cycle") throw error; }
if (!evaluateCondition({ field: "status", operator: "equals", value: "ready" }, { status: "ready" }) || evaluateCondition({ field: "status", operator: "equals", value: "ready" }, { status: "pending" })) throw new Error("workflow condition failed");
try { evaluateCondition({ field: "status", operator: "contains", value: "ready" }, { status: "ready" }); throw new Error("unsupported condition accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "unsupported condition") throw error; }
const approvals = createApprovalStore();
approvals.request("run-1", "admin");
if (approvals.isApproved("run-1") || !approvals.approve("run-1", "operator") || !approvals.isApproved("run-1")) throw new Error("approval flow failed");
if (approvals.approve("run-1", "operator")) throw new Error("duplicate approval accepted");
const workflowCommits = new InvocationCommitStore();
let workflowAttempts = 0;
const workflowResult = await runIdempotentWorkflow(workflowCommits, "workflow-1", async () => { workflowAttempts++; if (workflowAttempts < 2) throw new Error("retryable workflow failure"); return "done"; }, 3);
const duplicateWorkflowResult = await runIdempotentWorkflow(workflowCommits, "workflow-1", async () => { workflowAttempts++; return "wrong"; }, 3);
if (workflowResult !== "done" || duplicateWorkflowResult !== "done" || workflowAttempts !== 2) throw new Error("workflow retry/idempotency failed");
const safeStatus = redactRunStatus({ state: "running", token: "secret", nested: { password: "hidden", ok: true } });
if (JSON.stringify(safeStatus) !== JSON.stringify({ state: "running", token: "[redacted]", nested: { password: "[redacted]", ok: true } })) throw new Error("run status secret redaction failed");
