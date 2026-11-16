"use client";

import { useMemo, useRef, useState } from "react";
import type { KeyboardEvent } from "react";

type Completion = { label: string; insert: string; detail: string };

const completions: Completion[] = [
  {
    label: "async function",
    insert: "async function handler(input) {\n  \n}",
    detail: "Function declaration",
  },
  { label: "console.log", insert: "console.log(value);", detail: "Write a diagnostic log" },
  { label: "JSON.stringify", insert: "JSON.stringify(value)", detail: "Serialize a value" },
  { label: "return", insert: "return value;", detail: "Return from the function" },
  {
    label: "try / catch",
    insert: "try {\n  \n} catch (error) {\n  console.error(error);\n}",
    detail: "Error boundary",
  },
  { label: "fetch", insert: 'fetch(url, { method: "POST" })', detail: "Make an HTTP request" },
];

function escapeHtml(value: string) {
  return value.replace(
    /[&<>"']/g,
    (character) =>
      ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[character] ??
      character,
  );
}

function highlight(source: string) {
  return escapeHtml(source)
    .replace(/(\/\/.*)$/gm, '<span class="token-comment">$1</span>')
    .replace(/(&quot;.*?&quot;|&#39;.*?&#39;|`.*?`)/g, '<span class="token-string">$1</span>')
    .replace(
      /\b(export|default|async|await|function|return|const|let|if|else|try|catch|throw|new|import|from)\b/g,
      '<span class="token-keyword">$1</span>',
    )
    .replace(/\b(console|JSON|Math|Promise|fetch)\b/g, '<span class="token-global">$1</span>')
    .replace(/\b(true|false|null|undefined)\b/g, '<span class="token-value">$1</span>');
}

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
  const editorRef = useRef<HTMLTextAreaElement>(null);
  const [showCompletions, setShowCompletions] = useState(false);
  const [query, setQuery] = useState("");
  const lines = useMemo(() => source.split("\n"), [source]);
  const visibleCompletions = completions.filter(
    (completion) => query.length < 2 || completion.label.includes(query.toLowerCase()),
  );
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

  function insertCompletion(completion: Completion) {
    const editor = editorRef.current;
    if (!editor) return;
    const start = editor.selectionStart;
    onChange(`${source.slice(0, start)}${completion.insert}${source.slice(editor.selectionEnd)}`);
    setShowCompletions(false);
    requestAnimationFrame(() => {
      editor.focus();
      const cursor = start + completion.insert.length;
      editor.setSelectionRange(cursor, cursor);
    });
  }

  function handleKeyDown(event: KeyboardEvent<HTMLTextAreaElement>) {
    if ((event.ctrlKey || event.metaKey) && event.code === "Space") {
      event.preventDefault();
      setQuery("");
      setShowCompletions(true);
      return;
    }
    if (event.key === "Escape") {
      setShowCompletions(false);
      return;
    }
    if (event.key === "Tab") {
      event.preventDefault();
      const editor = event.currentTarget;
      const start = editor.selectionStart;
      onChange(`${source.slice(0, start)}  ${source.slice(editor.selectionEnd)}`);
      requestAnimationFrame(() => editor.setSelectionRange(start + 2, start + 2));
      return;
    }
    const pairs: Record<string, string> = { "(": ")", "[": "]", "{": "}", '"': '"', "'": "'" };
    if (
      pairs[event.key] &&
      event.currentTarget.selectionStart === event.currentTarget.selectionEnd
    ) {
      const editor = event.currentTarget;
      const cursor = editor.selectionStart;
      if (source[cursor] === pairs[event.key]) {
        event.preventDefault();
        editor.setSelectionRange(cursor + 1, cursor + 1);
        return;
      }
      event.preventDefault();
      onChange(`${source.slice(0, cursor)}${event.key}${pairs[event.key]}${source.slice(cursor)}`);
      requestAnimationFrame(() => editor.setSelectionRange(cursor + 1, cursor + 1));
    }
  }

  return (
    <section className="rich-editor" aria-label="TypeScript function editor">
      <header className="editor-toolbar">
        <span>
          <i className="editor-dot" /> TypeScript
        </span>
        <small>⌘/Ctrl + Space for completions · Tab to indent</small>
        {onValidate && (
          <button type="button" disabled={validating} onClick={onValidate}>
            {validating ? "Checking…" : "Check syntax"}
          </button>
        )}
        <button
          type="button"
          onClick={() => {
            setQuery("");
            setShowCompletions((visible) => !visible);
          }}
        >
          Completions
        </button>
      </header>
      <div className="editor-body">
        <div className="editor-gutter">
          {lines.map((_, index) => (
            <span key={index}>{String(index + 1).padStart(2, "0")}</span>
          ))}
        </div>
        <div className="editor-code">
          <pre aria-hidden="true" dangerouslySetInnerHTML={{ __html: `${highlight(source)}\n` }} />
          <textarea
            ref={editorRef}
            aria-label="Function source code"
            value={source}
            onChange={(event) => {
              onChange(event.target.value);
              setQuery(
                event.target.value.slice(0, event.target.selectionStart).split(/\s/).pop() ?? "",
              );
            }}
            onKeyDown={handleKeyDown}
            onScroll={(event) => {
              const pre = event.currentTarget.previousElementSibling as HTMLElement | null;
              if (pre) {
                pre.scrollTop = event.currentTarget.scrollTop;
                pre.scrollLeft = event.currentTarget.scrollLeft;
              }
            }}
            spellCheck={false}
          />
        </div>
      </div>
      {showCompletions && (
        <div className="editor-completion-menu" role="listbox">
          <div className="completion-heading">
            Suggestions <kbd>Esc</kbd>
          </div>
          {visibleCompletions.map((completion) => (
            <button
              type="button"
              role="option"
              key={completion.label}
              onClick={() => insertCompletion(completion)}
            >
              <strong>{completion.label}</strong>
              <span>{completion.detail}</span>
            </button>
          ))}
          {!visibleCompletions.length && <small className="muted">No matching completions.</small>}
        </div>
      )}
      {(localDiagnostic || validationMessage) && (
        <div
          className={`editor-diagnostic ${validationMessage && !localDiagnostic ? "valid" : "invalid"}`}
          role="status"
        >
          <strong>{validationMessage && !localDiagnostic ? "✓ Validated" : "Syntax check"}</strong>
          <span>{validationMessage || localDiagnostic}</span>
        </div>
      )}
      <footer className="editor-status">
        <span>TypeScript · {lines.length} lines</span>
        <span>{source.length} characters</span>
      </footer>
    </section>
  );
}
