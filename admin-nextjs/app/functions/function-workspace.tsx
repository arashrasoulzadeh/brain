"use client";

import { useId, useState, type ReactNode } from "react";

type WorkspaceProps = {
  tab: string;
  name: string;
  version: string;
  dirty: boolean;
  busy: boolean;
  validating: boolean;
  validVersion: boolean;
  canPublish: boolean;
  message: string;
  messageError: boolean;
  onValidate: () => void;
  onSave: () => void;
  onPublish: () => void;
  onRestore: () => void;
  editor: (expanded: boolean) => ReactNode;
  settings: ReactNode;
  runner: ReactNode;
};

export function FunctionWorkspace({ tab, name, version, dirty, busy, validating, validVersion, canPublish, message, messageError, onValidate, onSave, onPublish, onRestore, editor, settings, runner }: WorkspaceProps) {
  const [inspectorOpen, setInspectorOpen] = useState(false);
  const [consoleOpen, setConsoleOpen] = useState(false);
  const [expanded, setExpanded] = useState(false);
  const inspectorID = useId();
  const consoleID = useId();
  const editing = tab === "editor";
  const testing = tab === "test";
  const showInspector = inspectorOpen && !expanded;
  const locked = busy || validating;
  return <section className="function-workspace" aria-label="Function development workspace" hidden={!editing && !testing}>
    <div hidden={!editing}>
      <div className="function-workspace-toolbar" role="group" aria-label="Source actions">
        <div className="function-workspace-file">
          <span className="function-workspace-file-icon" aria-hidden="true">JS</span>
          <div><strong>{name}.js</strong><small role="status">{dirty ? "Unsaved changes" : "Saved draft"} · v{version || "—"}</small></div>
        </div>
        <div className="function-workspace-actions">
          <button type="button" className="secondary" disabled={locked} onClick={onValidate}>{validating ? "Validating…" : "Validate"}</button>
          <button type="button" className="secondary" disabled={locked || !dirty || !validVersion} onClick={onSave}>{busy ? "Working…" : "Save changes"}</button>
          <button type="button" disabled={locked || !canPublish || !validVersion} onClick={onPublish}>Publish</button>
        </div>
      </div>
      {message && <p className="function-workspace-notice notice" role={messageError ? "alert" : "status"}>{message}</p>}
      <div className="function-workspace-tools" role="group" aria-label="Workspace layout">
        <button type="button" className="secondary" aria-expanded={consoleOpen} aria-controls={consoleID} onClick={() => setConsoleOpen(value => !value)}>{consoleOpen ? "Hide test console" : "Test console"}</button>
        <button type="button" className="secondary" aria-expanded={showInspector} aria-controls={inspectorID} onClick={() => { if (expanded) { setExpanded(false); setInspectorOpen(true); } else setInspectorOpen(value => !value); }}>{showInspector ? "Hide settings" : "Settings"}</button>
        <button type="button" className="secondary" aria-pressed={expanded} onClick={() => setExpanded(value => !value)}>{expanded ? "Exit focus mode" : "Focus mode"}</button>
        <button type="button" className="secondary function-workspace-restore" disabled={!dirty || locked} onClick={onRestore}>Restore saved changes</button>
      </div>
      <div className={"function-workspace-grid" + (showInspector ? " has-inspector" : "")}>
        <div className="function-workspace-source">{editor(expanded)}</div>
        <aside id={inspectorID} className="function-workspace-inspector" aria-label="Function settings" hidden={!showInspector}>{settings}</aside>
      </div>
    </div>
    <div id={consoleID} className={editing ? "function-workspace-console" : ""} hidden={!testing && !(editing && consoleOpen)}>
      {runner}
    </div>
  </section>;
}
