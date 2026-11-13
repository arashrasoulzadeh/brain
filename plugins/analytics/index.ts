import { pluginRequest, validateProjectId } from "../sdk-typescript/src/index.js";

export type AnalyticsEvent = {
  projectId: string;
  name: string;
  fields: Record<string, string | number | boolean>;
};

export type EventSink = { append(event: AnalyticsEvent): void };

export function capture(event: AnalyticsEvent, sink: EventSink, allowedFields: readonly string[] = []) {
  validateEvent(event, allowedFields);
  sink.append({ projectId: event.projectId, name: event.name, fields: { ...event.fields } });
  return { captured: true, name: event.name };
}

export function captureRequest(projectId: string, pluginId: string, version: string, event: Omit<AnalyticsEvent, "projectId">) {
  validateEvent({ projectId, ...event });
  return pluginRequest("analytics", projectId, pluginId, version, "analytics.capture", event, event.name);
}

export function validateEvent(event: AnalyticsEvent, allowedFields: readonly string[] = []): void {
  validateProjectId(event.projectId);
  if (!/^[a-zA-Z][a-zA-Z0-9_.-]{0,63}$/.test(event.name)) throw new Error("invalid event name");
  const keys = Object.keys(event.fields);
  if (keys.length > 20 || keys.some((key) => !/^[a-zA-Z][a-zA-Z0-9_]{0,31}$/.test(key))) throw new Error("invalid event fields");
  if (allowedFields.length > 0 && keys.some((key) => !allowedFields.includes(key))) throw new Error("unapproved event field");
  if (keys.some((key) => event.fields[key] === undefined)) throw new Error("invalid event field value");
}
