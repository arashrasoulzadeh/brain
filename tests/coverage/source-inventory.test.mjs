import assert from "node:assert/strict";
import { test } from "node:test";
import { hasExecutableSource, validateReport } from "./source-inventory.mjs";

test("coverage classification uses emitted code, never a filename convention", () => {
  for (const text of [
    "export type Name = string;",
    "export interface Shape { a: string }",
    "import type { Other } from './other.js'; export type Name = Other;",
    "export {};",
    "// documentation only",
  ]) assert.equal(hasExecutableSource("types.ts", text), false, text);
  for (const text of [
    "export const name = 'x';",
    "export function read() { return 1; }",
    "import './side-effect.js';",
    "export { read } from './other.js';",
    "export enum Status { Good, Bad }",
    "export default function App() { return <p>Hi</p>; }",
  ]) assert.equal(hasExecutableSource("types.tsx", text), true, text);
  assert.equal(hasExecutableSource("anything.d.ts", "declare const a: number;"), false);
  assert.equal(hasExecutableSource("empty.mjs", ""), true);
});

test("coverage inventory rejects missing and unexpected runtime paths", () => {
  const expected = { runtime: ["plugins/read.ts"], erased: ["plugins/types.ts"] };
  validateReport({ total: {}, "plugins/read.ts": {} }, expected);
  assert.throws(() => validateReport({ total: {} }, expected), /missing/);
  assert.throws(() => validateReport({ "plugins/read.ts": {}, "plugins/types.ts": {} }, expected), /Unexpected/);
});
