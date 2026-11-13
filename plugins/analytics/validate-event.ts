import { validateProjectId } from "../sdk-typescript/src/index.js";
import type { AnalyticsEvent } from "./types.js";

export function validateEvent(event: AnalyticsEvent, allowedFields: readonly string[] = []): void {
  validateProjectId(event.projectId);
  if (!/^[a-zA-Z][a-zA-Z0-9_.-]{0,63}$/.test(event.name)) throw new Error("invalid event name");
  const fieldNames = Object.keys(event.fields);
  if (fieldNames.length > 20 || fieldNames.some((fieldName) => !/^[a-zA-Z][a-zA-Z0-9_]{0,31}$/.test(fieldName))) throw new Error("invalid event fields");
  if (allowedFields.length > 0 && fieldNames.some((fieldName) => !allowedFields.includes(fieldName))) throw new Error("unapproved event field");
  if (fieldNames.some((fieldName) => event.fields[fieldName] === undefined)) throw new Error("invalid event field value");
}
