// Pure helpers are shared by manifest workers and sandboxed user functions.
const registry = new Map([
  ["string.trim", (value) => { if (typeof value !== "string") throw new Error("expected string"); return value.trim(); }],
  ["string.nonempty", (value) => { if (typeof value !== "string" || !value.trim()) throw new Error("expected nonempty string"); return value; }],
  ["string.lower", (value) => { if (typeof value !== "string") throw new Error("expected string"); return value.toLowerCase(); }],
  ["string.upper", (value) => { if (typeof value !== "string") throw new Error("expected string"); return value.toUpperCase(); }],
  ["number.integer", (value) => { if (!Number.isSafeInteger(value)) throw new Error("expected integer"); return value; }],
  ["boolean.require", (value) => { if (typeof value !== "boolean") throw new Error("expected boolean"); return value; }],
  ["value.identity", (value) => value],
]);

export function hasHelper(name) {
  return registry.has(name);
}

export function callHelper(name, input) {
  const helper = registry.get(name);
  if (!helper) throw new Error("unknown pure helper");
  return helper(input);
}

export function helperNames() {
  return Object.freeze([...registry.keys()].sort());
}
