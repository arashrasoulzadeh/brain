export function loadDeclaredInputs(
  inputs: Readonly<Record<string, unknown>>,
  declarations: readonly string[],
): Readonly<Record<string, unknown>> {
  if (
    !inputs ||
    typeof inputs !== "object" ||
    declarations.some((name) => !/^[A-Za-z_][A-Za-z0-9_]*$/.test(name))
  )
    throw new Error("invalid input declaration");
  const loaded: Record<string, unknown> = {};
  for (const name of declarations) {
    if (!Object.prototype.hasOwnProperty.call(inputs, name))
      throw new Error("declared input missing");
    loaded[name] = inputs[name];
  }
  return loaded;
}
