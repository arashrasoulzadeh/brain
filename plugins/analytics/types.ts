export type AnalyticsEvent = {
  projectId: string;
  name: string;
  fields: Record<string, string | number | boolean>;
};

export type EventSink = { append(event: AnalyticsEvent): void };
