import { APPROVED_SDK_IMPORTS, ENTRY_PACKAGE_FIXTURE, functionRequest, PROJECT_HELPER_FIXTURE, validateSDKImports } from "../../plugins/functions/index.js";
validateSDKImports(APPROVED_SDK_IMPORTS);
const request = functionRequest("dev", "functions", "0.1.0", "functions.invoke", { name: "hello" });
if (request.params.projectId !== "dev" || request.params.operation !== "functions.invoke") throw new Error("functions request contract failed");
try { validateSDKImports(["brain/sdk", "os"]); throw new Error("unsupported SDK import accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "unsupported SDK import") throw error; }
if (!ENTRY_PACKAGE_FIXTURE.includes("package main")) throw new Error("entry package fixture contract failed");
if (!PROJECT_HELPER_FIXTURE.includes("package helper")) throw new Error("project helper fixture contract failed");
