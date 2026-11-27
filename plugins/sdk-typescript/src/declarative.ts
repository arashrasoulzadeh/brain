export type ModelSchema = Readonly<Record<string, "string" | "integer" | "boolean">>;

export function validateModel(schema: ModelSchema, input: unknown): Readonly<Record<string, string | number | boolean>> {
  if (!input || typeof input !== "object" || Array.isArray(input)) throw new Error("expected model object");
  const source = input as Record<string, unknown>;
  const fields = Object.entries(schema);
  if (fields.length > 64 || Object.keys(source).length !== fields.length) throw new Error("invalid model fields");
  const result: Record<string, string | number | boolean> = Object.create(null);
  for (const [key, kind] of fields) {
    if (!/^[A-Za-z_][A-Za-z0-9_]*$/.test(key) || ["constructor", "prototype", "__proto__"].includes(key) || !Object.hasOwn(source, key)) throw new Error("invalid model field");
    const value = source[key];
    if (kind === "string" && typeof value === "string") result[key] = value;
    else if (kind === "integer" && typeof value === "number" && Number.isSafeInteger(value)) result[key] = value;
    else if (kind === "boolean" && typeof value === "boolean") result[key] = value;
    else throw new Error("invalid model field type");
  }
  return Object.freeze(result);
}

export type StateMachine = Readonly<{
  initial: string;
  transitions: Readonly<Record<string, Readonly<{ from: string; to: string }>>>;
}>;

// Pure transition calculation. Persistence and compare-and-swap remain host responsibilities.
export function transitionState(machine: StateMachine, current: string, action: string): string {
  if (!Object.hasOwn(machine.transitions, action)) throw new Error("unknown transition");
  const transition = machine.transitions[action];
  if (transition.from !== current || !transition.to) throw new Error("invalid transition");
  return transition.to;
}
