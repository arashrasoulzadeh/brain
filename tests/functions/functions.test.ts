import { APPROVED_SDK_IMPORTS, ENTRY_PACKAGE_FIXTURE, functionRequest, PROJECT_HELPER_FIXTURE, requireSDKImports, validateSDKImports, wasmBuildSettings } from "../../plugins/functions/index.js";
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
