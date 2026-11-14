export function normalizePhoneNumber(value: string): string {
  const compact = value.replace(/[\s().-]/g, "");
  const normalized = compact.startsWith("00") ? `+${compact.slice(2)}` : compact;
  if (!/^\+[1-9][0-9]{7,14}$/.test(normalized)) throw new Error("invalid phone number");
  return normalized;
}
