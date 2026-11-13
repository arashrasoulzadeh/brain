export function requireExplicitPublication(publish: boolean): void {
  if (publish !== true) throw new Error("explicit publication required");
}
