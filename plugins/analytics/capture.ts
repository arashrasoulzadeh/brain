import type { AnalyticsEvent, EventSink } from "./types.js";
import { validateEvent } from "./validate-event.js";
import { analyticsRouteSink } from "./analytics-route-sink.js";

export function capture(
  event: AnalyticsEvent,
  sink: EventSink = analyticsRouteSink,
  allowedFields: readonly string[] = [],
) {
  validateEvent(event, allowedFields);
  sink.append({ projectId: event.projectId, name: event.name, fields: { ...event.fields } });
  return { captured: true, name: event.name };
}
