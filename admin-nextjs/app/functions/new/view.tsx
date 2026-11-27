"use client";
import { useEffect, useRef, useState, type FormEvent } from "react";
import { usePathname, useRouter } from "next/navigation";
import Link from "next/link";
import { adminUIContext, projectUIURL, readAdminSession } from "../../lib/admin-api";
import { AdminDialog, AdminProgress } from "../../components/admin-dialog";
import { AdminPageFrame } from "../../components/admin-chrome";
import { FunctionEditor } from "../function-editor";
import { byteLength, checkedFunctionDraft, functionFormError, functionTemplates, MAX_SOURCE_BYTES, normalizePermissions, resolveFunctionScope, sendFunctionCommand, type FunctionTemplate } from "../../lib/function-studio";

export default function NewFunctionPage() {
  const { namespace } = adminUIContext(usePathname());
  return <AdminPageFrame activePage="functions" title="Create function" requiredPluginID="functions"><CreateFunction key={namespace ?? ""} namespace={namespace} /></AdminPageFrame>;
}

function CreateFunction({ namespace }: { namespace?: string }) {
  const router = useRouter();
  const [name, setName] = useState("");
  const [template, setTemplate] = useState<FunctionTemplate>("echo");
  const [pendingTemplate, setPendingTemplate] = useState<FunctionTemplate>();
  const [source, setSource] = useState<string>(functionTemplates.echo.source);
  const [version, setVersion] = useState("1");
  const [publicRoute, setPublicRoute] = useState(false);
  const [permissions, setPermissions] = useState("functions.invoke");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const [confirm, setConfirm] = useState(false);
  const lifetime = useRef<AbortController | null>(null);
  const pending = useRef(false);
  const formError = functionFormError(name, source, version);
  const changed = Boolean(name || source !== functionTemplates.echo.source || version !== "1" || publicRoute || permissions !== "functions.invoke");
  useEffect(() => {
    const controller = new AbortController();
    lifetime.current = controller;
    return () => { controller.abort(); if (lifetime.current === controller) lifetime.current = null; };
  }, []);
  useEffect(() => {
    if (!changed) return;
    function warn(event: BeforeUnloadEvent) { event.preventDefault(); event.returnValue = ""; }
    window.addEventListener("beforeunload", warn);
    return () => window.removeEventListener("beforeunload", warn);
  }, [changed]);
  function applyTemplate(next: FunctionTemplate) {
    setTemplate(next); setSource(functionTemplates[next].source); setPendingTemplate(undefined);
  }
  async function create() {
    const controller = lifetime.current;
    if (!controller || pending.current || formError) return;
    const token = readAdminSession();
    pending.current = true; setBusy(true); setError("");
    try {
      const scope = await resolveFunctionScope(token, namespace, controller.signal);
      if (controller.signal.aborted || readAdminSession() !== token) return;
      checkedFunctionDraft(await sendFunctionCommand(scope, { operation: "functions.create", name, source, function_version: Number(version), public: publicRoute, permissions: normalizePermissions(permissions) }, controller.signal), scope, name);
      if (!controller.signal.aborted && readAdminSession() === token) router.push(projectUIURL(scope.project, "functions", name));
    } catch (reason) { if (!controller.signal.aborted && readAdminSession() === token) setError(String(reason)); }
    finally { pending.current = false; if (!controller.signal.aborted && readAdminSession() === token) setBusy(false); }
  }
  function submit(event: FormEvent) {
    event.preventDefault();
    if (formError) { setError(formError); return; }
    if (publicRoute) setConfirm(true);
    else void create();
  }
  return <div className="fn-workspace">
    <div className="fn-heading"><div><p className="fn-kicker">NEW FUNCTION</p><h2>Start with a module</h2><p className="muted">Choose a starter, customize the source and create an unpublished function.</p></div>{namespace && <Link href={projectUIURL(namespace, "functions")}>Back to functions</Link>}</div>
    <div className="fn-templates" aria-label="Function templates">{(Object.keys(functionTemplates) as FunctionTemplate[]).map(key => <button key={key} type="button" className={"fn-template " + (key === template ? "active" : "")} aria-pressed={key === template} disabled={busy} onClick={() => { if (key === template) return; if (source !== functionTemplates[template].source) setPendingTemplate(key); else applyTemplate(key); }}><strong>{functionTemplates[key].title}</strong><span>{key === "health" ? "GET · Health check" : key === "transform" ? "POST · SDK helper" : "POST · Request and response"}</span></button>)}</div>
    <form className="fn-studio-grid" onSubmit={submit}>
      <section className="fn-panel fn-code"><div className="fn-panel-heading"><h3>Source code</h3><span className="muted">{byteLength(source).toLocaleString()} / {MAX_SOURCE_BYTES.toLocaleString()} bytes</span></div><FunctionEditor source={source} onChange={setSource} readOnly={busy} /><p className="muted">Use JavaScript-compatible module source. The HTTP method is declared in the exported config. The default handler must return a JSON object, array or null.</p></section>
      <section className="fn-panel"><h3>Function details</h3><label>Function name<input value={name} onChange={event => setName(event.target.value)} disabled={busy} placeholder="hello-world" required pattern="[A-Za-z][A-Za-z0-9_\-]{0,63}" maxLength={64} /></label><small className="muted">Start with a letter; use letters, numbers, underscores or hyphens.</small><label>Version<input type="number" min="1" step="1" value={version} disabled={busy} onChange={event => setVersion(event.target.value)} /></label><label>Permissions (comma-separated)<input value={permissions} disabled={busy} onChange={event => setPermissions(event.target.value)} /></label><label className="checkbox-label"><input type="checkbox" checked={publicRoute} disabled={busy} onChange={event => setPublicRoute(event.target.checked)} /> Public route enabled</label><p className="fn-warning">Creating a function does not publish an endpoint. Keep visibility private unless public access is intended.</p><div className="fn-actions"><button type="submit" disabled={busy || !!formError}>Create function</button></div>{name && formError && <p role="alert" className="fn-error">{formError}</p>}{error && <p role="alert" className="fn-error">{error}</p>}{busy && <AdminProgress label="Creating function…" />}<p className="muted">After creation, use Code &amp; test for host validation and JSON execution before publication.</p></section>
    </form>
    {pendingTemplate && <AdminDialog title="Replace edited source" scope={namespace || "Function"} onClose={() => setPendingTemplate(undefined)}><p>Switching templates will replace your local source edits.</p><button type="button" onClick={() => applyTemplate(pendingTemplate)}>Replace source</button></AdminDialog>}
    {confirm && <AdminDialog title="Create with public visibility" scope={(namespace || "") + " · " + name} busy={busy} onClose={() => setConfirm(false)}><p>Create this unpublished function with public visibility configured? Caller policy still applies after publication.</p><button type="button" disabled={busy} onClick={() => void create()}>Confirm creation</button>{error && <p role="alert">{error}</p>}</AdminDialog>}
  </div>;
}
