import { pluginRequest, validateProjectId, type PluginContext, type Observation } from "../sdk-typescript/src/index.js";

export type AnalyticsEvent = {
  projectId: string;
  name: string;
  fields: Record<string, string | number | boolean>;
};

export type EventSink = { append(event: AnalyticsEvent): void };

export function validateEvent(event: AnalyticsEvent, allowedFields: readonly string[] = []): void {
  validateProjectId(event.projectId);
  if (!/^[a-zA-Z][a-zA-Z0-9_.-]{0,63}$/.test(event.name)) throw new Error("invalid event name");
  const fieldNames = Object.keys(event.fields);
  if (
    fieldNames.length > 20 ||
    fieldNames.some((fieldName) => !/^[a-zA-Z][a-zA-Z0-9_]{0,31}$/.test(fieldName))
  )
    throw new Error("invalid event fields");
  if (
    allowedFields.length > 0 &&
    fieldNames.some((fieldName) => !allowedFields.includes(fieldName))
  )
    throw new Error("unapproved event field");
  if (fieldNames.some((fieldName) => event.fields[fieldName] === undefined))
    throw new Error("invalid event field value");
  if (
    fieldNames.some((fieldName) => {
      const value = event.fields[fieldName];
      return (
        (typeof value === "number" && !Number.isFinite(value)) ||
        (typeof value !== "string" && typeof value !== "number" && typeof value !== "boolean") ||
        (typeof value === "string" && value.length > 1024)
      );
    })
  )
    throw new Error("invalid event field value");
}

export class AnalyticsRouteSink implements EventSink {
  readonly events: AnalyticsEvent[] = [];
  append(event: AnalyticsEvent): void {
    this.events.push({ ...event, fields: { ...event.fields } });
  }
}

export const analyticsRouteSink = new AnalyticsRouteSink();

export function capture(
  event: AnalyticsEvent,
  sink: EventSink = analyticsRouteSink,
  allowedFields: readonly string[] = [],
) {
  validateEvent(event, allowedFields);
  sink.append({ projectId: event.projectId, name: event.name, fields: { ...event.fields } });
  return { captured: true, name: event.name };
}

/** Production entry point: use the host-bound project, never a payload identity. */
export function captureForProject(context: PluginContext, input: unknown) {
  const event = input as Partial<AnalyticsEvent> | null;
  if (!event || (event.projectId !== undefined && event.projectId !== context.project.id) ||
      typeof event.name !== "string" || !event.fields || typeof event.fields !== "object")
    throw new Error("invalid analytics event or project context");
  return capture({ projectId: context.project.id, name: event.name, fields: event.fields });
}

export function observeRoomConnected(context: PluginContext, input: unknown) {
  const observation = input as Observation<{ roomId: string; reconnected: boolean }>;
  if (!observation || observation.projectId !== context.project.id || observation.pluginId !== "rooms" || observation.event !== "connected")
    throw new Error("invalid room observation");
  return captureForProject(context, { name: "rooms.connected", fields: observation.payload });
}

export function captureRequest(
  projectId: string,
  pluginId: string,
  version: string,
  event: Omit<AnalyticsEvent, "projectId">,
) {
  validateEvent({ projectId, ...event });
  return pluginRequest(
    "analytics",
    projectId,
    pluginId,
    version,
    "analytics.capture",
    event,
    event.name,
  );
}
