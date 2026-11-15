import type { AnalyticsEvent, EventSink } from "./types.js";

export class AnalyticsRouteSink implements EventSink {
  readonly events: AnalyticsEvent[] = [];
  append(event: AnalyticsEvent): void {
    this.events.push({ ...event, fields: { ...event.fields } });
  }
}

export const analyticsRouteSink = new AnalyticsRouteSink();
