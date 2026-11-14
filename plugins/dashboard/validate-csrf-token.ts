export function validateCsrfToken(expected: string, provided: string): void {
  if (!expected || !provided || expected.length !== provided.length)
    throw new Error("csrf validation failed");
  let difference = 0;
  for (let index = 0; index < expected.length; index += 1)
    difference |= expected.charCodeAt(index) ^ provided.charCodeAt(index);
  if (difference !== 0) throw new Error("csrf validation failed");
}
