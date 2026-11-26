export function configurationPatch(schema: Record<string, string>, values: Record<string, string>): Record<string, string | number | boolean> {
  const result: Record<string, string | number | boolean> = Object.create(null);
  for (const [key, value] of Object.entries(values)) {
    if (!Object.hasOwn(schema, key)) throw new Error("Unknown configuration field: " + key);
    switch (schema[key]) {
      case "string": result[key] = value; break;
      case "int":
        if (!/^-?\d+$/.test(value) || !Number.isSafeInteger(Number(value))) throw new Error(key + " must be an integer");
        result[key] = Number(value); break;
      case "bool":
        if (value !== "true" && value !== "false") throw new Error(key + " must be true or false");
        result[key] = value === "true"; break;
      default: throw new Error("Unsupported configuration type: " + schema[key]);
    }
  }
  return result;
}
