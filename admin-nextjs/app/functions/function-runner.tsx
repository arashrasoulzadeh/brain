"use client";

import { useEffect, useRef, useState } from "react";
import { adminHeaders, readAdminSession, projectAdminURL, ADMIN_SESSION_CHANGED } from "../lib/admin-api";
import { functionInput, functionNameError } from "../lib/function-studio";
import { AdminProgress } from "../components/admin-dialog";

type RunResult = { status: "success" | "error"; message: string; source: string; input: string };
export function FunctionRunner({ project, name, source, disabled = false, onComplete }: {
  project: string; name: string; source: string; disabled?: boolean; onComplete?: () => void;
}) {
  const [input, setInput] = useState("{}");
  const [result, setResult] = useState<RunResult>();
  const [busy, setBusy] = useState(false);
  const pending = useRef<AbortController | null>(null);
  useEffect(() => {
    const clear = () => { pending.current?.abort(); pending.current = null; setBusy(false); setResult(undefined); };
    window.addEventListener(ADMIN_SESSION_CHANGED, clear);
    return () => { window.removeEventListener(ADMIN_SESSION_CHANGED, clear); pending.current?.abort(); pending.current = null; };
  }, [project, name]);

  async function run() {
    if (pending.current || disabled) return;
    const token = readAdminSession();
    const snapshot = { source, input };
    let payload: unknown;
    try {
      if (!token || !project) throw new Error("Select an active project and sign in.");
      const error = functionNameError(name);
      if (error) throw new Error(error);
      if (!source.trim()) throw new Error("Function source is required.");
      payload = functionInput(input);
    } catch (reason) {
      setResult({ status: "error", message: String(reason instanceof Error ? reason.message : reason), ...snapshot });
      return;
    }
    const controller = new AbortController();
    pending.current = controller;
    setBusy(true); setResult(undefined);
    const current = () => !controller.signal.aborted && readAdminSession() === token;
    try {
      const response = await fetch(projectAdminURL(project, "functions"), {
        method: "POST", headers: adminHeaders(token, { "content-type": "application/json" }), signal: controller.signal,
        body: JSON.stringify({ operation: "functions.run", project_id: project, name: name.trim(), source, input: payload }),
      });
      const body = await response.text();
      if (!current()) return;
      if (!response.ok) throw new Error(body || "Execution failed (" + response.status + ").");
      const data = JSON.parse(body) as { valid?: boolean; result?: unknown };
      if (data.valid !== true || !Object.prototype.hasOwnProperty.call(data, "result")) throw new Error("Unexpected execution response.");
      setResult({ status: "success", message: JSON.stringify(data.result, null, 2) ?? "null", ...snapshot });
    } catch (reason) {
      if (current()) setResult({ status: "error", message: reason instanceof Error ? reason.message : String(reason), ...snapshot });
    } finally {
      if (pending.current === controller) {
        pending.current = null;
        if (current()) { setBusy(false); onComplete?.(); }
      }
    }
  }

  return <section className="section function-runner" aria-label="Sandbox execution">
    <div className="section-heading"><h2>Sandbox execution</h2><button type="button" disabled={busy || disabled || !project} onClick={() => void run()}>Run test</button></div>
    <p className="muted">Runs the current source with JSON input using your operator authorization. It does not publish or save source; execution may create logs. This is not a public caller test.</p>
    <div className="function-runner-grid">
      <label>JSON input<textarea rows={10} value={input} disabled={busy || disabled} onChange={event => setInput(event.target.value)} spellCheck={false} /></label>
      <div className="function-runner-result" aria-live="polite">
        <h3>Execution result</h3>
        {busy ? <AdminProgress label="Executing in sandbox…" /> : result ? <>
          {(result.source !== source || result.input !== input) && <p className="muted">Result is from a previous source or input.</p>}
          <p role={result.status === "error" ? "alert" : "status"}>{result.status === "success" ? "Execution succeeded" : "Execution failed"}</p>
          <pre>{result.message}</pre>
        </> : <p className="muted">Run a test to view its response.</p>}
      </div>
    </div>
  </section>;
}
