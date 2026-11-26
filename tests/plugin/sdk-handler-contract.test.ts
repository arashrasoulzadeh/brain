import assert from "node:assert/strict";
import { test } from "node:test";
import {
  createFunctionHandler,
  createHostHandler,
  createRouteHandler,
  request,
} from "../../plugins/sdk-typescript/src/index.js";
import type { PluginRoute } from "../../plugins/sdk-typescript/src/index.js";

const context = { projectId: "dev", invocationId: "one" };
const routes: PluginRoute[] = [{ operation: "read", function: "read", capability: "read" }];
const rpc = request("one", {
  protocol: 1, projectId: "dev", pluginId: "sample", version: "1.0.0", operation: "read",
});

test("SDK handlers share manifest route validation", () => {
  for (const invalidRoutes of [[], [{ ...routes[0], capability: "" }]]) {
    const handlers = [
      createRouteHandler(invalidRoutes),
      createHostHandler({ invoke: () => { throw new Error("host must not run"); } }, invalidRoutes, context),
      createFunctionHandler({ read: () => { throw new Error("function must not run"); } }, invalidRoutes, context),
    ];
    for (const handler of handlers) {
      assert.throws(() => handler(rpc), /operation not declared in plugin manifest|invalid plugin route/);
    }
  }
});

test("host and function handlers reject a mismatched project before dispatch", () => {
  const wrongContext = { ...context, projectId: "other" };
  const handlers = [
    createHostHandler({ invoke: () => { throw new Error("host must not run"); } }, routes, wrongContext),
    createFunctionHandler({ read: () => { throw new Error("function must not run"); } }, routes, wrongContext),
  ];
  for (const handler of handlers) {
    assert.throws(() => handler(rpc), /host project context mismatch/);
  }
});

test("function handler only resolves explicitly implemented functions", () => {
  const handler = createFunctionHandler({}, [{ ...routes[0], function: "toString" }], context);
  assert.throws(() => handler(rpc), /route function not implemented/);
});
