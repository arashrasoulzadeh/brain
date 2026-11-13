import { validateIdentifier } from "../sdk-typescript/src/index.js";

export function validateValue(column: string, value: unknown): void {
  validateIdentifier(column);
  if (value === undefined || typeof value === "function" || (typeof value === "string" && value.length > 4096)) throw new Error("invalid value");
}
