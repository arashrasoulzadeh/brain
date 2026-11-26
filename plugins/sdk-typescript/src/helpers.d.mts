export type PureHelperName = "string.trim" | "string.nonempty" | "string.lower" | "string.upper" | "number.integer" | "boolean.require" | "value.identity";
export function hasHelper(name: unknown): name is PureHelperName;
export function callHelper(name: PureHelperName, input: unknown): unknown;
export function helperNames(): readonly PureHelperName[];
