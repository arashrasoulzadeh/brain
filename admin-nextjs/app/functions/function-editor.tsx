"use client";

import dynamic from "next/dynamic";
import { useMemo } from "react";

const MonacoEditor = dynamic(() => import("@monaco-editor/react"), { ssr: false });

export function FunctionEditor({
  source,
  onChange,
  onValidate,
  validating = false,
  validationMessage = "",
}: {
  source: string;
  onChange: (source: string) => void;
  onValidate?: () => void;
  validating?: boolean;
  validationMessage?: string;
}) {
  const localDiagnostic = useMemo(() => {
    const stack: string[] = [];
    const pairs: Record<string, string> = { "(": ")", "[": "]", "{": "}" };
    for (const character of source) {
      if (pairs[character]) stack.push(pairs[character]);
      else if (Object.values(pairs).includes(character) && stack.pop() !== character)
        return "Unbalanced brackets or braces";
    }
    if (stack.length) return "Unclosed bracket or brace";
    if (!/\bexport\s+default\b/.test(source)) return "A default exported handler is required";
    if (/\b(import|require|process|child_process|fs|net|fetch|Deno)\b/.test(source))
      return "Host law: imports, network access, process, and filesystem access are not allowed";
    if (
      !/export\s+const\s+config\s*=\s*\{[\s\S]*\bmethod\s*:\s*["'](GET|POST|PUT|PATCH)["']/.test(
        source,
      )
    )
      return "Host law: export config with method GET, POST, PUT, or PATCH";
    return "";
  }, [source]);

  return (
    <section className="rich-editor monaco-editor-shell" aria-label="Function source code editor">
      <header className="editor-toolbar">
        <span>Function source</span>
        <small>Monaco Editor · TypeScript</small>
        {onValidate && (
          <button type="button" disabled={validating} onClick={onValidate}>
            {validating ? "Checking…" : "Check syntax"}
          </button>
        )}
      </header>
      <MonacoEditor
        height="520px"
        language="typescript"
        theme="vs"
        value={source}
        onChange={(value) => onChange(value ?? "")}
        options={{
          automaticLayout: true,
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
      {(localDiagnostic || validationMessage) && (
        <div
          className={`editor-diagnostic ${validationMessage && !localDiagnostic ? "valid" : "invalid"}`}
          role="status"
        >
          <strong>{validationMessage && !localDiagnostic ? "✓ Validated" : "Syntax check"}</strong>
          <span>{validationMessage || localDiagnostic}</span>
        </div>
      )}
    </section>
  );
}
