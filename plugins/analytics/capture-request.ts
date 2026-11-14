import { pluginRequest } from "../sdk-typescript/src/index.js";
import type { AnalyticsEvent } from "./types.js";
import { validateEvent } from "./validate-event.js";

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
