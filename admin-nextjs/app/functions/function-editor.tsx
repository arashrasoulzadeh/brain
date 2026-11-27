"use client";

import dynamic from "next/dynamic";
import { useState } from "react";
import type { OnValidate } from "@monaco-editor/react";

const MonacoEditor = dynamic(() => import("@monaco-editor/react"), { ssr: false });

export function FunctionEditor({
  source,
  onChange,
  onValidate,
  validating = false,
  validationMessage = "",
  validationOK = false,
  readOnly = false,
}: {
  source: string;
  onChange: (source: string) => void;
  onValidate?: () => void;
  validating?: boolean;
  validationMessage?: string;
  validationOK?: boolean;
  readOnly?: boolean;
}) {
  const [diagnostics, setDiagnostics] = useState<{ line: number; column: number; message: string }[]>([]);
  const showDiagnostics: OnValidate = markers => setDiagnostics(markers.filter(marker => marker.severity === 8).map(marker => ({ line: marker.startLineNumber, column: marker.startColumn, message: marker.message })));

  return (
    <section className="monaco-function-editor" aria-label="Function source code editor">
      <div className="monaco-editor-actions">
        <span>Function source · Monaco</span>
        {onValidate && (
          <button type="button" disabled={validating || readOnly} onClick={onValidate}>
            {validating ? "Checking…" : "Check syntax"}
          </button>
        )}
      </div>
      <MonacoEditor
        height="520px"
        language="typescript"
        theme="vs"
        value={source}
        onChange={(value) => { setDiagnostics([]); onChange(value ?? ""); }}
        onValidate={showDiagnostics}
        options={{
          automaticLayout: true,
          readOnly,
          minimap: { enabled: false },
          lineNumbers: "on",
          folding: true,
          wordWrap: "on",
          tabSize: 2,
          insertSpaces: true,
          scrollBeyondLastLine: false,
          fontSize: 12,
          lineHeight: 18,
          suggestOnTriggerCharacters: true,
          quickSuggestions: true,
        }}
      />
      <p className="muted">Editor diagnostics are advisory. Host validation checks the supported function contract and sandbox restrictions.</p>
      {!!diagnostics.length && <details className="editor-diagnostic"><summary>{diagnostics.length} editor diagnostics</summary><ul>{diagnostics.slice(0, 20).map((item, index) => <li key={index}>Line {item.line}:{item.column} · {item.message}</li>)}</ul></details>}
      {validationMessage && (
        <div
          className={`editor-diagnostic ${validationMessage && validationOK ? "valid" : "invalid"}`}
          role="status"
        >
          <strong>{validationOK ? "✓ Host validation passed" : "Host validation failed"}</strong>
          <span>{validationMessage}</span>
        </div>
      )}
    </section>
  );
}
