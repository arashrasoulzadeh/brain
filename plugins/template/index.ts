import { pluginRequest, requireProjectInput, type PluginContext, type Observation } from "../sdk-typescript/src/index.js";

export function invoke(projectId: string, pluginId: string, version: string, operation: string) {
  return pluginRequest(pluginId, projectId, pluginId, version, operation);
}

export function example(context: PluginContext, _input: unknown) {
  context.plugins.observe.fire("exampleCompleted", {});
  return { projectId: context.project.id, accepted: true };
}

/** A subscriber does not import the publisher or receive caller-supplied identity. */
export function observeRoomConnected(context: PluginContext, input: unknown) {
  requireProjectInput(context, input);
  const event = input as Observation<{ roomId: string; reconnected: boolean }>;
  if (event.projectId !== context.project.id || event.pluginId !== "rooms" || event.event !== "connected" ||
      !event.payload || typeof event.payload.roomId !== "string" || typeof event.payload.reconnected !== "boolean")
    throw new Error("invalid room observation");
  return { recorded: true };
}
