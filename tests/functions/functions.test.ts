import { APPROVED_SDK_IMPORTS, APPROVED_WASI_IMPORTS, capDiagnostics, compilerEnvironment, createInvocationRequest, diagnoseSource, DisposableCompilerWorkspace, ENTRY_PACKAGE_FIXTURE, executeFixture, FunctionDraftStore, functionRequest, InvocationEffects, parseInvocationResult, PROJECT_HELPER_FIXTURE, rejectSandboxEscapeAttempt, requireSDKImports, runWithTimeout, SharedPackageStore, validateBuildOverrides, validateCompilerEnvironment, validateCompilerLimits, validateFunctionLimits, validateInvocationResult, validateLinearMemory, validateSandboxCapabilities, validateSDKImports, validateSourceProject, validateStartupDuration, validateWasiImports, wasmBuildSettings } from "../../plugins/functions/index.js";
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
const effects = new InvocationEffects(); const failed = effects.begin(); failed.write({ id: "failed" }); failed.commit(false);
if (effects.list().length !== 0) throw new Error("failed invocation wrote effects");
const succeeded = effects.begin(); succeeded.write({ id: "ok" }); succeeded.commit(true);
if (effects.list().length !== 1) throw new Error("successful invocation effects missing");
const drafts = new FunctionDraftStore(); drafts.save("dev", "hello", "package main", 1000);
if (drafts.get("dev", "hello")?.source !== "package main" || drafts.get("prod", "hello")) throw new Error("function draft isolation failed");
const packages = new SharedPackageStore(); packages.save("dev", "helper", "package helper");
if (packages.load("dev", "helper") !== "package helper") throw new Error("shared package load failed");
try { packages.load("prod", "helper"); throw new Error("cross-project package loaded"); } catch (error) { if (!(error instanceof Error) || error.message !== "shared package not found") throw error; }
if (diagnoseSource("package main\nimport \"os\"").length !== 1 || diagnoseSource("package main\nimport \"os\"")[0].line !== 2) throw new Error("source diagnostics failed");
if (!executeFixture(() => ({ success: true, output: "ok", diagnostics: [] })).success) throw new Error("fixture execution failed");
const fixtureFailure = executeFixture(() => { throw new Error("fixture boom"); });
if (fixtureFailure.success || fixtureFailure.error !== "fixture boom") throw new Error("fixture failure handling failed");
