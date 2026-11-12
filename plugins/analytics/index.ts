import { request } from "../sdk-typescript/src/index.js";

export type AnalyticsEvent = {
  projectId: string;
  name: string;
  fields: Record<string, string | number | boolean>;
};

export function captureRequest(projectId: string, pluginId: string, version: string, event: Omit<AnalyticsEvent, "projectId">) {
  validateEvent({ projectId, ...event });
  return request(`analytics-capture-${projectId}-${event.name}`, { protocol: 1, projectId, pluginId, version, operation: "analytics.capture", payload: event });
}

export function validateEvent(event: AnalyticsEvent): void {
  if (!event.projectId || !/^[a-zA-Z0-9][a-zA-Z0-9_-]*$/.test(event.projectId)) throw new Error("invalid project id");
  if (!/^[a-zA-Z][a-zA-Z0-9_.-]{0,63}$/.test(event.name)) throw new Error("invalid event name");
  const keys = Object.keys(event.fields);
  if (keys.length > 20 || keys.some((key) => !/^[a-zA-Z][a-zA-Z0-9_]{0,31}$/.test(key))) throw new Error("invalid event fields");
  if (keys.some((key) => event.fields[key] === undefined)) throw new Error("invalid event field value");
}
