"use client";

import dynamic from "next/dynamic";
import { useEffect, useRef, useState } from "react";
import type { OnMount, OnValidate } from "@monaco-editor/react";

const MonacoEditor = dynamic(() => import("@monaco-editor/react"), { ssr: false, loading: () => <p role="status" className="editor-loading">Loading source editor…</p> });
type MonacoInstance = Parameters<OnMount>[0];
type EditorTheme = "vs" | "vs-dark";

export function FunctionEditor({
  source, onChange, onValidate, validating = false, validationMessage = "", validationOK = false,
  readOnly = false, expanded = false, onSave, saveEnabled = false, initialTheme = "vs",
}: {
  source: string;
  onChange: (source: string) => void;
  onValidate?: () => void;
  validating?: boolean;
  validationMessage?: string;
  validationOK?: boolean;
  readOnly?: boolean;
  expanded?: boolean;
  onSave?: () => void;
  saveEnabled?: boolean;
  initialTheme?: EditorTheme;
}) {
  const [diagnostics, setDiagnostics] = useState<{ line: number; column: number; message: string }[]>([]);
  const [theme, setTheme] = useState<EditorTheme>(initialTheme);
  const [wrap, setWrap] = useState(false);
  const [ready, setReady] = useState(false);
  const [position, setPosition] = useState({ lineNumber: 1, column: 1 });
  const [toolMessage, setToolMessage] = useState("");
  const editorRef = useRef<MonacoInstance | null>(null);
  const saveRef = useRef({ onSave, saveEnabled, readOnly });
  const cursorSubscription = useRef<{ dispose: () => void } | null>(null);

  useEffect(() => { saveRef.current = { onSave, saveEnabled, readOnly }; }, [onSave, saveEnabled, readOnly]);
  useEffect(() => () => { cursorSubscription.current?.dispose(); editorRef.current = null; }, []);

  const showDiagnostics: OnValidate = markers => setDiagnostics(markers.filter(marker => marker.severity === 8).map(marker => ({ line: marker.startLineNumber, column: marker.startColumn, message: marker.message })));
  const mountEditor: OnMount = (editor, monaco) => {
    editorRef.current = editor;
    setReady(true);
    cursorSubscription.current?.dispose();
    cursorSubscription.current = editor.onDidChangeCursorPosition(event => setPosition(event.position));
    if (onSave) editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyS, () => {
      const current = saveRef.current;
      if (current.saveEnabled && !current.readOnly) current.onSave?.();
    });
  };

  async function editorAction(action: "actions.find" | "editor.action.formatDocument") {
    const editor = editorRef.current;
    const command = editor?.getAction(action);
    setToolMessage("");
    if (!editor || !command || !command.isSupported()) {
      setToolMessage("This editor action is not available.");
      return;
    }
    try { editor.focus(); await command.run(); }
    catch { setToolMessage("Could not complete the editor action."); }
  }

  return <section className={"monaco-function-editor function-code-editor" + (theme === "vs-dark" ? " is-dark" : "")} aria-label="Function source code editor">
    <div className="monaco-editor-actions">
      <span className="function-code-language">JavaScript <small>· Monaco</small></span>
      <div className="function-code-tools" role="group" aria-label="Editor tools">
        <button type="button" disabled={!ready} onClick={() => void editorAction("actions.find")}>Find</button>
        <button type="button" disabled={!ready || readOnly} onClick={() => void editorAction("editor.action.formatDocument")}>Format</button>
        <button type="button" aria-pressed={wrap} onClick={() => setWrap(value => !value)}>Word wrap</button>
        <button type="button" aria-label="Dark editor theme" aria-pressed={theme === "vs-dark"} onClick={() => setTheme(value => value === "vs" ? "vs-dark" : "vs")}>Dark theme</button>
        {onValidate && <button type="button" disabled={validating || readOnly} onClick={onValidate}>{validating ? "Checking…" : "Check syntax"}</button>}
      </div>
    </div>
    <MonacoEditor
      height={expanded ? "clamp(520px, 76dvh, 1200px)" : "clamp(400px, 58dvh, 900px)"}
      language="javascript"
      theme={theme}
      value={source}
      onMount={mountEditor}
      onChange={value => { setDiagnostics([]); onChange(value ?? ""); }}
      onValidate={showDiagnostics}
      loading={<p role="status" className="editor-loading">Loading Monaco…</p>}
      options={{
        ariaLabel: "Function JavaScript source",
        automaticLayout: true,
        readOnly,
        minimap: { enabled: false },
        lineNumbers: "on",
        folding: true,
        wordWrap: wrap ? "on" : "off",
        tabSize: 2,
        insertSpaces: true,
        scrollBeyondLastLine: false,
        fontSize: 13,
        lineHeight: 21,
        suggestOnTriggerCharacters: true,
        quickSuggestions: true,
      }}
    />
    <div className="function-code-status">
      <span>{ready ? "Ln " + position.lineNumber + ", Col " + position.column : "Editor loading"}</span>
      <span>{source.split("\n").length} lines</span>
      <span>Spaces: 2</span>
      <span>UTF-8</span>
      <span>{diagnostics.length ? diagnostics.length + " editor errors" : "No editor errors reported"}</span>
      {onSave && <span className="function-code-shortcut">Ctrl/⌘ + S to save</span>}
    </div>
    {toolMessage && <p className="function-code-note" role="status">{toolMessage}</p>}
    {!!diagnostics.length && <details className="editor-diagnostic"><summary>{diagnostics.length} editor diagnostics</summary><ul>{diagnostics.slice(0, 20).map((item, index) => <li key={index}><button type="button" className="secondary" disabled={!ready} onClick={() => { editorRef.current?.setPosition({ lineNumber: item.line, column: item.column }); editorRef.current?.revealLineInCenter(item.line); editorRef.current?.focus(); }}>Line {item.line}:{item.column}</button> {item.message}</li>)}</ul></details>}
    {validationMessage && <div className={"editor-diagnostic " + (validationOK ? "valid" : "invalid")} role={validationOK ? "status" : "alert"}><strong>{validationOK ? "Host preflight passed" : "Host preflight failed"}</strong><span>{validationMessage}</span></div>}
    <p className="function-code-note">Editor diagnostics are advisory. Host preflight checks the source contract; sandbox execution is a separate test. JavaScript-compatible modules only.</p>
  </section>;
}
