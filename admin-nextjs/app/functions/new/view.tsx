"use client";

import { useEffect, useRef, useState, type FormEvent } from "react";
import { usePathname, useRouter } from "next/navigation";
import Link from "next/link";
import { adminHeaders, readAdminSession, preferredAdminProject, projectAdminURL, systemAdminURL, projectUIURL, adminUIContext, ADMIN_SESSION_CHANGED, ADMIN_PROJECT_CHANGED } from "../../lib/admin-api";
import { AdminDialog, AdminProgress } from "../../components/admin-dialog";
import { AdminPageFrame } from "../../components/admin-chrome";
import { FunctionEditor } from "../function-editor";
import { FunctionRunner } from "../function-runner";
import { functionTemplate, functionMethods, functionNameError, functionVersion, functionPermissions, type FunctionTemplate, type FunctionMethod } from "../../lib/function-studio";

export default function NewFunctionPage() {
  const { namespace } = adminUIContext(usePathname());
  const [epoch, setEpoch] = useState(0);
  useEffect(() => {
    const changed = () => setEpoch(value => value + 1);
    window.addEventListener(ADMIN_SESSION_CHANGED, changed);
    window.addEventListener(ADMIN_PROJECT_CHANGED, changed);
    return () => { window.removeEventListener(ADMIN_SESSION_CHANGED, changed); window.removeEventListener(ADMIN_PROJECT_CHANGED, changed); };
  }, []);
  return <AdminPageFrame key={(namespace ?? "") + ":" + epoch} activePage="functions" title="Create function" requiredPluginID="functions"><NewFunctionStudio namespace={namespace} /></AdminPageFrame>;
}

function NewFunctionStudio({ namespace }: { namespace?: string }) {
  const router = useRouter();
  const [project, setProject] = useState("");
  const [loadError, setLoadError] = useState("");
  const [reload, setReload] = useState(0);
  const [name, setName] = useState("");
  const [source, setSource] = useState(() => functionTemplate("echo"));
  const [version, setVersion] = useState("1");
  const [publicRoute, setPublicRoute] = useState(false);
  const [permissions, setPermissions] = useState("functions.invoke");
  const [template, setTemplate] = useState<FunctionTemplate>("echo");
  const [method, setMethod] = useState<FunctionMethod>("POST");
  const [replacePending, setReplacePending] = useState(false);
  const [message, setMessage] = useState("");
  const [busy, setBusy] = useState(false);
  const [validationMessage, setValidationMessage] = useState("");
  const [validationOK, setValidationOK] = useState(false);
  const [validating, setValidating] = useState(false);
  const pending = useRef(false);
  const lifetime = useRef<AbortController | null>(null);
  const dirty = Boolean(name || source !== functionTemplate("echo") || version !== "1" || publicRoute || permissions !== "functions.invoke");

  useEffect(() => {
    const controller = new AbortController();
    lifetime.current = controller;
    const token = readAdminSession();
    setProject(""); setLoadError("");
    void (async () => {
      if (!token || !namespace) throw new Error("Select an active project and sign in");
      const response = await fetch(systemAdminURL("projects"), { headers: adminHeaders(token), signal: controller.signal });
      if (!response.ok) throw new Error("Could not load projects");
      const selected = preferredAdminProject((await response.json() as { ID: string; Slug?: string; Active: boolean }[] | null) ?? [], namespace)?.ID;
      if (!selected) throw new Error("Project unavailable");
      if (!controller.signal.aborted && readAdminSession() === token) setProject(selected);
    })().catch(reason => { if (!controller.signal.aborted && readAdminSession() === token) setLoadError(String(reason)); });
    return () => { controller.abort(); if (lifetime.current === controller) lifetime.current = null; };
  }, [namespace, reload]);

  useEffect(() => {
    if (!dirty) return;
    const warn = (event: BeforeUnloadEvent) => { event.preventDefault(); event.returnValue = ""; };
    window.addEventListener("beforeunload", warn);
    return () => window.removeEventListener("beforeunload", warn);
  }, [dirty]);

  function changeSource(value: string) { setSource(value); setValidationMessage(""); setValidationOK(false); }

  async function validate() {
    if (pending.current) return;
    const token = readAdminSession();
    const signal = lifetime.current?.signal;
    const error = functionNameError(name);
    if (!project || !token || !signal || signal.aborted || error) { setValidationMessage(error || "Select an active project and sign in"); setValidationOK(false); return; }
    pending.current = true; setValidating(true); setValidationMessage(""); setValidationOK(false);
    try {
      const response = await fetch(projectAdminURL(project, "functions"), {
        method: "POST", headers: adminHeaders(token, { "content-type": "application/json" }), signal,
        body: JSON.stringify({ operation: "functions.test", project_id: project, name: name.trim(), source }),
      });
      const body = await response.text();
      if (signal.aborted || readAdminSession() !== token) return;
      setValidationOK(response.ok);
      setValidationMessage(response.ok ? "Source passed host preflight. Run a sandbox test to execute it." : body || "Validation failed.");
    } catch (reason) { if (!signal.aborted && readAdminSession() === token) setValidationMessage(String(reason)); }
    finally { pending.current = false; if (!signal.aborted) setValidating(false); }
  }

  async function save(event: FormEvent) {
    event.preventDefault();
    if (pending.current) return;
    const token = readAdminSession();
    const signal = lifetime.current?.signal;
    const number = functionVersion(version);
    const error = functionNameError(name);
    if (!project || !token || !signal || signal.aborted || error || number === undefined || !source.trim()) {
      setMessage(error || (number === undefined ? "Version must be a positive integer" : !source.trim() ? "Source is required" : "Select an active project and sign in")); return;
    }
    pending.current = true; setBusy(true); setMessage("");
    try {
      const response = await fetch(projectAdminURL(project, "functions"), { method: "POST", headers: adminHeaders(token, { "content-type": "application/json" }), signal,
        body: JSON.stringify({ operation: "functions.create", project_id: project, name: name.trim(), source, function_version: number, public: publicRoute, permissions: functionPermissions(permissions) }) });
      if (!response.ok) throw new Error(await response.text() || "Creation failed.");
      if (!signal.aborted && readAdminSession() === token) router.push(projectUIURL(project, "functions", name.trim()));
    } catch (reason) { if (!signal.aborted && readAdminSession() === token) setMessage(String(reason)); }
    finally { pending.current = false; if (!signal.aborted) setBusy(false); }
  }

  if (loadError) return <section className="section"><p role="alert">{loadError}</p><button type="button" className="secondary" onClick={() => setReload(value => value + 1)}>Retry</button></section>;
  if (!project) return <AdminProgress label="Loading project…" />;
  const locked = busy || validating;
  return <div className="function-catalog">
    <div className="section-heading"><Link href={projectUIURL(project, "functions")} onClick={event => { if (dirty && !window.confirm("Leave without creating this function?")) event.preventDefault(); }}>Back to functions</Link><span className="muted">{project} · New draft</span></div>
    <form className="function-studio" onSubmit={save}>
      <section className="section"><h2>Function configuration</h2>
        <div className="function-settings">
          <label>Function name<input disabled={locked} value={name} onChange={event => { setName(event.target.value); setValidationMessage(""); setValidationOK(false); }} required maxLength={64} placeholder="hello" /></label>
          <label>Version<input disabled={locked} type="number" min="1" step="1" value={version} onChange={event => setVersion(event.target.value)} required /></label>
          <label>Permissions<input disabled={locked} value={permissions} onChange={event => setPermissions(event.target.value)} /></label>
          <label className="checkbox-label"><input disabled={locked} type="checkbox" checked={publicRoute} onChange={event => setPublicRoute(event.target.checked)} /> Public call route</label>
        </div>
        <p className="muted">Use a unique name. Visibility is configuration; actual access is enforced by host policy. Creating a draft does not publish it.</p>
      </section>
      <section className="section">
        <div className="section-heading"><h2>Source</h2><div className="function-catalog-actions">
          <label>Starter template<select disabled={locked} value={template} onChange={event => setTemplate(event.target.value as FunctionTemplate)}><option value="echo">Echo JSON</option><option value="health">Health check</option><option value="transform">Array transform</option></select></label>
          <label>Template method<select disabled={locked} value={method} onChange={event => setMethod(event.target.value as FunctionMethod)}>{functionMethods.map(value => <option key={value}>{value}</option>)}</select></label>
          <button type="button" className="secondary" disabled={locked} onClick={() => setReplacePending(true)}>Apply template</button>
        </div></div>
        <p className="muted">The method selector applies only to the starter template; the exported source config is authoritative.</p>
        <FunctionEditor source={source} onChange={changeSource} readOnly={locked} onValidate={() => void validate()} validating={validating} validationMessage={validationMessage} validationOK={validationOK} />
      </section>
      <div className="playground-actions"><button type="submit" disabled={locked}>Create function</button>{message && <small role="alert">{message}</small>}{busy && <AdminProgress label="Creating function…" />}</div>
    </form>
    {replacePending && <AdminDialog title="Replace function source" scope={project + " · New draft"} onClose={() => setReplacePending(false)}><p>Applying a starter template replaces all current source.</p><button type="button" onClick={() => { changeSource(functionTemplate(template, method)); setReplacePending(false); }}>Replace source</button></AdminDialog>}
    <FunctionRunner key={project + ":" + name} project={project} name={name} source={source} disabled={locked} />
  </div>;
}
