export function validateIdentifier(
  identifier: string,
  errorMessage = "invalid identifier",
): string {
  if (!/^[a-zA-Z_][a-zA-Z0-9_]*$/.test(identifier)) throw new Error(errorMessage);
  return identifier;
}
