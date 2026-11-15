"use client";

import { useMemo } from "react";

export function FunctionEditor({ source, onChange }: { source: string; onChange: (source: string) => void }) {
  const lines = useMemo(() => source.split("\n").length, [source]);
  return <div className="rich-editor"><div className="editor-gutter">{Array.from({ length: lines }, (_, index) => <span key={index}>{index + 1}</span>)}</div><textarea aria-label="Function source code" value={source} onChange={(event) => onChange(event.target.value)} spellCheck={false} /><div className="editor-completions" aria-label="Completion suggestions"><code>export default async function handler(input)</code><code>return input</code><code>console.log()</code><code>JSON.stringify()</code></div></div>;
}
