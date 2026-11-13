import { APPROVED_SDK_IMPORTS, functionRequest, validateSDKImports } from "../../plugins/functions/index.js";
validateSDKImports(APPROVED_SDK_IMPORTS);
const request = functionRequest("dev", "functions", "0.1.0", "functions.invoke", { name: "hello" });
if (request.params.projectId !== "dev" || request.params.operation !== "functions.invoke") throw new Error("functions request contract failed");
try { validateSDKImports(["brain/sdk", "os"]); throw new Error("unsupported SDK import accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "unsupported SDK import") throw error; }
