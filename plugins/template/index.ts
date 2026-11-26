import { pluginRequest, type PluginContext } from "../sdk-typescript/src/index.js";

export function invoke(projectId: string, pluginId: string, version: string, operation: string) {
  return pluginRequest(pluginId, projectId, pluginId, version, operation);
}

export function example(context: PluginContext, _input: unknown) {
  context.plugins.observe.fire("exampleCompleted", {});
  return { projectId: context.project.id, accepted: true };
}
