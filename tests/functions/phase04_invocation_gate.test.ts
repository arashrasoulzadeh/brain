import { createRecordChangeTriggerStore, createScheduleStore, invokeDirectHttps, parseInvocationResult } from "../../plugins/functions/index.js";

const direct = await invokeDirectHttps("https://functions.example/invoke", { projectId: "dev" }, async () => new Response(JSON.stringify({ success: true, output: { path: "direct" }, diagnostics: [] }), { status: 200 }));
if (!direct.success || (direct.output as { path: string }).path !== "direct") throw new Error("direct invocation gate failed");
if (!parseInvocationResult({ success: true, output: { path: "fixture" }, diagnostics: [] }).success) throw new Error("fixture invocation gate failed");
const schedules = createScheduleStore();
schedules.add("dev", "sync", "0 * * * *", 1000);
if (!schedules.due(1000).length || !schedules.claim("dev", "sync", "worker", 1000, 500)) throw new Error("scheduled invocation gate failed");
const triggers = createRecordChangeTriggerStore();
triggers.add("dev", "profiles", "profile.changed", { functionId: "sync", version: 1 });
if (triggers.dispatch("dev", "profiles", "profile.changed", false).length || triggers.dispatch("dev", "profiles", "profile.changed", true).length !== 1) throw new Error("record-change invocation gate failed");
