"use client";
import { adminHeaders, readAdminSession, preferredAdminProject, projectAdminURL, systemAdminURL, projectUIURL, systemUIURL, adminUIContext, ADMIN_SESSION_CHANGED, ADMIN_PROJECT_CHANGED } from "../../lib/admin-api";
import { useEffect, useRef, useState } from "react";
import { useParams, usePathname } from "next/navigation";
import Link from "next/link";
import { AdminDialog, AdminProgress } from "../../components/admin-dialog";
import { AdminPageFrame } from "../../components/admin-chrome";
import { FunctionEditor } from "../function-editor";
import { FunctionRunner } from "../function-runner";
import { FunctionWorkspace } from "../function-workspace";
import { functionPermissions, functionVersion } from "../../lib/function-studio";
import { filterFunctionLogs, functionLogMetrics, type FunctionLog as Log, type LogOutcome } from "../../lib/function-inspection";
type Draft = {
  project_id: string;
  name: string;
  source: string;
  version?: number;
  public?: boolean;
  permissions?: string[];
  method?: string;
  published_at?: string;
  active_publication?: number;
  published_versions?: Record<string, { version: number; method: string; source: string; digest: string }>;
};
export default function FunctionDetailPage() {
  const params = useParams<{ name: string }>();
  const { namespace: requestedProject } = adminUIContext(usePathname());
  const name = decodeURIComponent(params.name ?? "");
  const [epoch, setEpoch] = useState(0);
  useEffect(() => {
    const changed = () => setEpoch(value => value + 1);
    window.addEventListener(ADMIN_SESSION_CHANGED, changed);
    window.addEventListener(ADMIN_PROJECT_CHANGED, changed);
    return () => { window.removeEventListener(ADMIN_SESSION_CHANGED, changed); window.removeEventListener(ADMIN_PROJECT_CHANGED, changed); };
  }, []);
  const scope = (requestedProject ?? "") + ":" + name + ":" + epoch;
  return <AdminPageFrame key={scope} activePage="functions" title={name} requiredPluginID="functions"><FunctionDetail key={scope} name={name} requestedProject={requestedProject} /></AdminPageFrame>;
}
function FunctionDetail({ name, requestedProject }: { name: string; requestedProject?: string }) {
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState("");
  const [logError, setLogError] = useState("");
  const [logsLoaded, setLogsLoaded] = useState(false);
  const [validationOK, setValidationOK] = useState(false);
  const [settingsPending, setSettingsPending] = useState<{ permissions: string[]; publicRoute: boolean }>();
  const [publishPending, setPublishPending] = useState(false);
  const [session, setSession] = useState("");
  const [project, setProject] = useState("");
  const [draft, setDraft] = useState<Draft | null>(null);
  const [savedSource, setSavedSource] = useState("");
  const [savedVersion, setSavedVersion] = useState(1);
  const [versionInput, setVersionInput] = useState("1");
  const [permissionInput, setPermissionInput] = useState("");
  const [reload, setReload] = useState(0);
  const [discardPending, setDiscardPending] = useState(false);
  const dirty = draft !== null && (draft.source !== savedSource || versionInput !== String(savedVersion));
  useEffect(() => {
    if (!dirty) return;
    function warn(event: BeforeUnloadEvent) { event.preventDefault(); event.returnValue = ""; }
    window.addEventListener("beforeunload", warn);
    return () => window.removeEventListener("beforeunload", warn);
  }, [dirty]);
  const [logs, setLogs] = useState<Log[]>([]);
  const [tab, setTab] = useState("editor");
  const [message, setMessage] = useState("");
  const [messageError, setMessageError] = useState(false);
  function report(message: string, error = false) { setMessage(message); setMessageError(error); }
  const [route, setRoute] = useState("");
  const [publishing, setPublishing] = useState(false);
  const [validating, setValidating] = useState(false);
  const [validationMessage, setValidationMessage] = useState("");
  const [logPage, setLogPage] = useState(0);
  const [logQuery, setLogQuery] = useState("");
  const [logOutcome, setLogOutcome] = useState<LogOutcome>("all");
  const [logRefresh, setLogRefresh] = useState(0);
  const [selectedLog, setSelectedLog] = useState<Log | null>(null);
  const lifetime = useRef<AbortController | null>(null);
  async function refresh(token: string, preserveSource?: string) {
    const signal = lifetime.current?.signal;
    if (!signal || signal.aborted) return;
    const options = { headers: adminHeaders(token), signal };
    const projectsResponse = await fetch(systemAdminURL("projects"), options);
    if (!projectsResponse.ok) throw new Error("Could not load projects");
    const projects = await projectsResponse.json() as { ID: string; Slug?: string; Active: boolean }[];
    const selected = preferredAdminProject(projects, requestedProject);
    const active = selected?.ID;
    if (!active) throw new Error("No active project selected");
    const draftsResponse = await fetch(projectAdminURL(active, "functions"), options);
    if (!draftsResponse.ok) throw new Error(await draftsResponse.text());
    const drafts = (await draftsResponse.json() as Draft[] | null) ?? [];
    const current = drafts.find(candidate => candidate.project_id === active && candidate.name === name) ?? null;
    if (signal.aborted || readAdminSession() !== token) return;
    setProject(active);
    setSavedSource(current?.source ?? "");
    setSavedVersion(current?.version ?? 1);
    setVersionInput(String(current?.version ?? 1));
    setDraft(current && preserveSource !== undefined ? { ...current, source: preserveSource } : current);
    const currentRoute = current?.published_at ? window.location.origin + "/" + encodeURIComponent(selected?.Slug || active) + "/functions/" + encodeURIComponent(name) : "";
    setRoute(currentRoute);
  }
  useEffect(() => {
    const controller = new AbortController();
    lifetime.current = controller;
    const token = readAdminSession();
    setSession(token); setLoading(true); setLoadError("");
    if (!token) { setLoadError("Sign in to continue"); setLoading(false); return () => controller.abort(); }
    void refresh(token).catch(reason => { if (!controller.signal.aborted) setLoadError(String(reason)); }).finally(() => { if (!controller.signal.aborted) setLoading(false); });
    return () => { controller.abort(); if (lifetime.current === controller) lifetime.current = null; };
  }, [name, requestedProject, reload]);
  useEffect(() => {
    if (!session || !project || !draft) return;
    const controller = new AbortController();
    let timer: ReturnType<typeof setTimeout>;
    async function poll() {
      try {
        const response = await fetch(projectAdminURL(project, "function-logs") + "?name=" + encodeURIComponent(name), { headers: adminHeaders(session, { "x-brain-quiet": "1" }), signal: controller.signal });
        if (!response.ok) throw new Error("Function logs unavailable");
        const rows = (await response.json() as Log[] | null) ?? [];
        if (!controller.signal.aborted && readAdminSession() === session) { setLogs(rows); setLogsLoaded(true); setLogError(""); setLogPage(page => Math.min(page, Math.max(0, Math.ceil(rows.length / 10) - 1))); }
      } catch (reason) { if (!controller.signal.aborted) { setLogs([]); setLogsLoaded(false); setSelectedLog(null); setLogError(String(reason)); } }
      finally { if (!controller.signal.aborted) timer = setTimeout(() => void poll(), 5000); }
    }
    void poll();
    return () => { controller.abort(); clearTimeout(timer); };
  }, [project, session, name, logRefresh, Boolean(draft)]);
  async function savePermissions(permissions: string[], publicRoute: boolean) {
    if (!draft || publishing || validating || readAdminSession() !== session || lifetime.current?.signal.aborted !== false) return;
    setPublishing(true); report("");
    try {
      const response = await fetch(projectAdminURL(project, "functions"), {
        method: "POST", headers: adminHeaders(session, { "content-type": "application/json" }), signal: lifetime.current?.signal,
        body: JSON.stringify({ operation: "functions.permissions", project_id: project, name: draft.name, permissions, public: publicRoute }),
      });
      if (!response.ok) throw new Error(await response.text());
      if (lifetime.current?.signal.aborted !== false || readAdminSession() !== session) return;
      setDraft(previous => previous ? { ...previous, permissions, public: publicRoute } : previous);
      setPermissionInput("");
      report("Settings saved."); setSettingsPending(undefined);
    } catch (reason) { report(String(reason), true); }
    finally { setPublishing(false); }
  }
  async function publish() {
    if (publishing || validating || !draft || !project || !session || readAdminSession() !== session || lifetime.current?.signal.aborted !== false) {
      report("Select an active project and sign in before publishing.", true);
      return;
    }
    const version = functionVersion(versionInput);
    if (version === undefined) { report("Version must be a positive integer.", true); return; }
    if (!draft.name.trim() || !draft.source.trim()) {
      report("Cannot publish: function name and source are required.", true);
      return;
    }
    setPublishing(true);
    report("Saving source and publishing…");
    try {
      const headers = adminHeaders(session, { "content-type": "application/json" });
      const update = await fetch(projectAdminURL(project, "functions"), {
        method: "POST",
        headers, signal: lifetime.current?.signal,
        body: JSON.stringify({
          operation: "functions.update",
          project_id: project,
          name: draft.name,
          source: draft.source,
          function_version: version,
        }),
      });
      if (!update.ok) {
        report((await update.text()) || `Source save failed (${update.status}).`, true);
        return;
      }
      if (lifetime.current?.signal.aborted !== false || readAdminSession() !== session) return;
      setSavedSource(draft.source); setSavedVersion(version);
      const response = await fetch(projectAdminURL(project, "functions"), {
        method: "POST",
        headers, signal: lifetime.current?.signal,
        body: JSON.stringify({
          operation: "functions.publish",
          project_id: project,
          name: draft.name,
          function_version: version,
          permissions: draft.permissions ?? [],
        }),
      });
      const body = await response.text();
      if (lifetime.current?.signal.aborted !== false || readAdminSession() !== session) return;
      if (!response.ok) {
        report("Source saved; publication failed: " + (body || String(response.status)), true);
        return;
      }
      const publication = JSON.parse(body) as { version: number };
      setSavedVersion(publication.version);
      setVersionInput(String(publication.version));
      report("Published version " + publication.version + ".");
      setPublishPending(false);
      await refresh(session).catch(reason => report("Published; refresh failed: " + String(reason), true));
    } catch (error) {
      report(error instanceof Error ? error.message : "Publish request failed.", true);
    } finally {
      setPublishing(false);
    }
  }
  async function validateSource() {
    if (validating || publishing) return;
    if (!draft || !project || !session || readAdminSession() !== session || lifetime.current?.signal.aborted !== false) {
      setValidationMessage("Select an active project and sign in before validating.");
      return;
    }
    setValidating(true);
    setValidationMessage(""); setValidationOK(false);
    try {
      const response = await fetch(projectAdminURL(project, "functions"), {
        method: "POST",
        headers: adminHeaders(session, { "content-type": "application/json" }), signal: lifetime.current?.signal,
        body: JSON.stringify({
          operation: "functions.test",
          project_id: project,
          name: draft.name,
          source: draft.source,
        }),
      });
      const body = await response.text();
      if (lifetime.current?.signal.aborted !== false || readAdminSession() !== session) return;
      if (!response.ok) {
        setValidationMessage(body || `Validation failed (${response.status}).`);
        return;
      }
      setValidationOK(true);
      setValidationMessage("Source passed host validation and is ready for sandbox compilation.");
    } catch (error) {
      setValidationMessage(error instanceof Error ? error.message : "Validation request failed.");
    } finally {
      setValidating(false);
    }
  }
  async function saveDraft() {
    if (publishing || validating) return;
    if (!draft || !project || !session || readAdminSession() !== session || lifetime.current?.signal.aborted !== false) {
      report("Select an active project and sign in before saving.", true);
      return;
    }
    const version = functionVersion(versionInput);
    if (version === undefined) { report("Version must be a positive integer.", true); return; }
    if (!draft.name.trim() || !draft.source.trim()) {
      report("Cannot save: function name and source are required.", true);
      return;
    }
    setPublishing(true);
    try {
      const response = await fetch(projectAdminURL(project, "functions"), {
        method: "POST",
        headers: adminHeaders(session, { "content-type": "application/json" }), signal: lifetime.current?.signal,
        body: JSON.stringify({
          operation: "functions.update",
          project_id: project,
          name: draft.name,
          source: draft.source,
          function_version: version,
        }),
      });
      if (lifetime.current?.signal.aborted !== false || readAdminSession() !== session) return;
      report(response.ok ? "Changes saved." : (await response.text() || "Source save failed."), !response.ok);
      if (response.ok) { setSavedSource(draft.source); setSavedVersion(version); await refresh(session).catch(reason => report("Source saved; refresh failed: " + String(reason), true)); }
    } catch (reason) { report(String(reason), true); } finally {
      setPublishing(false);
    }
  }
  if (loading) return <AdminProgress label="Loading function…" />;
  if (loadError) return <section className="section"><p role="alert">{loadError}</p><button type="button" className="secondary" onClick={() => setReload(value => value + 1)}>Retry</button></section>;
  if (!draft) return <><p role="status">Function not found.</p>{project && <Link href={projectUIURL(project, "functions")}>Back to functions</Link>}</>;
  const permissions = draft.permissions ?? [];
  const { successful, average, p95 } = functionLogMetrics(logs);
  const orderedLogs = filterFunctionLogs(logs, logQuery, logOutcome);
  const logPageCount = Math.max(1, Math.ceil(orderedLogs.length / 10));
  const currentLogPage = Math.min(logPage, logPageCount - 1);
  const visibleLogs = orderedLogs.slice(currentLogPage * 10, currentLogPage * 10 + 10);
  return (
    <>
      <div className="function-detail-heading">
        <Link href={project ? projectUIURL(project, "functions") : systemUIURL("projects")} onClick={event => { if (dirty && !window.confirm("Discard unsaved function changes?")) event.preventDefault(); }}>Back to functions</Link>
        <div className="function-detail-meta"><span>{project}</span><code>{draft.method ?? "—"}</code><span className={"status-pill " + (draft.published_at ? "is-success" : "is-inactive")}>{draft.published_at ? "Publication recorded" : "Unpublished"}</span><span>{draft.public ? "Public" : "Private"}</span></div>
      </div>
      <nav className="function-tabs function-detail-tabs" aria-label="Function workspace views">
        {[
          ["editor", "Editor"],
          ["test", "Test"],
          ["deployments", "Publication"],
          ["analytics", "Analytics"],
          ["logs", "Logs"],
        ].map(([value, label]) => (
          <button
            type="button"
            className={tab === value ? "active" : ""}
            aria-pressed={tab === value}
            onClick={() => setTab(value)}
            key={value}
          >
            {label}
          </button>
        ))}
      </nav>
      {discardPending && <AdminDialog title="Discard unsaved changes" scope={project + " · " + name} onClose={() => setDiscardPending(false)}><p>Restore the last saved source and version? Local edits will be lost.</p><button type="button" onClick={() => { setDraft(previous => previous ? { ...previous, source: savedSource } : previous); setVersionInput(String(savedVersion)); setValidationMessage(""); setValidationOK(false); setDiscardPending(false); }}>Discard edits</button></AdminDialog>}
      {publishPending && <AdminDialog title="Publish current version" scope={project + " · " + name} busy={publishing} onClose={() => setPublishPending(false)}><p>Save the current source and publish version {versionInput}? If this version already contains different published source, Brain will allocate the next unused version.</p><p>Publishing enables public route visibility on the current host record. Caller authorization is still enforced by host policy.</p><button type="button" disabled={publishing} onClick={() => void publish()}>Confirm publish</button>{publishing && <AdminProgress label="Saving and publishing…" />}{message && <p role={messageError ? "alert" : "status"}>{message}</p>}</AdminDialog>}
      {settingsPending && <AdminDialog title="Update function settings" scope={project + " · " + name} busy={publishing} onClose={() => setSettingsPending(undefined)}><p>{settingsPending.publicRoute ? "Public route enabled" : "Private route"} · Permissions: {settingsPending.permissions.join(", ") || "None"}</p><button type="button" disabled={publishing} onClick={() => void savePermissions(settingsPending.permissions, settingsPending.publicRoute)}>Confirm settings</button>{publishing && <AdminProgress label="Saving function settings…" />}{message && <p role={messageError ? "alert" : "status"}>{message}</p>}</AdminDialog>}
      {tab !== "editor" && message && <p role={messageError ? "alert" : "status"}>{message}</p>}
      {logError && (tab === "logs" || tab === "analytics") && <section><p role="alert">{logError}</p><button type="button" className="secondary" onClick={() => setLogRefresh(value => value + 1)}>Refresh logs</button></section>}
      <FunctionWorkspace
        tab={tab} name={name} version={versionInput} dirty={dirty} busy={publishing} validating={validating}
        validVersion={functionVersion(versionInput) !== undefined} canPublish={Boolean(draft.source.trim())}
        message={message} messageError={messageError}
        onValidate={() => void validateSource()} onSave={() => void saveDraft()}
        onPublish={() => { report(""); setPublishPending(true); }}
        onRestore={() => setDiscardPending(true)}
        editor={expanded => <FunctionEditor
          source={draft.source}
          onChange={source => { setDraft(previous => previous ? { ...previous, source } : previous); setValidationMessage(""); setValidationOK(false); report(""); }}
          readOnly={publishing || validating}
          expanded={expanded}
          initialTheme="vs-dark"
          onSave={() => void saveDraft()}
          saveEnabled={dirty && functionVersion(versionInput) !== undefined}
          validationMessage={validationMessage}
          validationOK={validationOK}
        />}
        settings={<div className="function-inspector-content">
          <h2>Function settings</h2>
          <label>Source version<input type="number" min="1" step="1" value={versionInput} disabled={publishing || validating} onChange={event => setVersionInput(event.target.value)} /></label>
          {functionVersion(versionInput) === undefined && <p role="alert">Version must be a positive integer.</p>}
          <p className="muted">Saved draft version: {savedVersion}. Published version: {draft.active_publication ?? "—"}. Publishing changed source automatically advances a conflicting version.</p>
          <div className="function-inspector-block"><h3>HTTP method</h3><code>{draft.method ?? "—"}</code><p className="muted">Saved method. Edit the exported config in source to change it.</p></div>
          <div className="function-inspector-block">
            <h3>Visibility</h3>
            <label className="checkbox-label"><input type="checkbox" checked={Boolean(draft.public)} disabled={publishing || validating} onChange={event => setSettingsPending({ permissions, publicRoute: event.target.checked })} /> Public route enabled</label>
            <p className="muted">Host access policy still controls callers. Publishing enables public visibility.</p>
            {route && <><small>Recorded route</small><code className="function-inspector-route">{route}</code></>}
          </div>
          <div className="function-inspector-block">
            <h3>Configured permissions <span className="muted">({permissions.length})</span></h3>
            {!permissions.length && <p className="muted">No configured permissions.</p>}
            {permissions.map(permission => <div className="function-inspector-permission" key={permission}><code>{permission}</code><button type="button" className="danger" aria-label={"Remove " + permission} disabled={publishing || validating} onClick={() => setSettingsPending({ permissions: permissions.filter(candidate => candidate !== permission), publicRoute: Boolean(draft.public) })}>Remove</button></div>)}
            <label>Additional permissions<input value={permissionInput} disabled={publishing || validating} onChange={event => setPermissionInput(event.target.value)} placeholder="Capability names, comma-separated" /></label>
            <button type="button" className="secondary" disabled={publishing || validating || !functionPermissions(permissionInput).some(permission => !permissions.includes(permission))} onClick={() => setSettingsPending({ permissions: [...new Set([...permissions, ...functionPermissions(permissionInput)])], publicRoute: Boolean(draft.public) })}>Add permissions</button>
            <button type="button" className="secondary" disabled={publishing || validating || permissions.includes("functions.invoke")} onClick={() => setSettingsPending({ permissions: [...permissions, "functions.invoke"], publicRoute: Boolean(draft.public) })}>Add invoke permission</button>
            <p className="muted">Stored permission names do not grant plugin access to sandbox source.</p>
          </div>
        </div>}
        runner={<FunctionRunner project={project} name={name} source={draft.source} disabled={publishing || validating} onComplete={() => setLogRefresh(value => value + 1)} />}
      />
      {tab === "deployments" && (
        <section className="section">
          <h2>Current publication</h2>
          <p className="muted">The published source and method are pinned independently of the editable draft. A new publication never overwrites an existing version.</p>
          {draft.published_at ? <div className="table-wrap"><table><tbody>
            <tr><th scope="row">Last published</th><td><time dateTime={draft.published_at}>{new Date(draft.published_at).toLocaleString()}</time></td></tr>
            <tr><th scope="row">Current saved source version</th><td>{savedVersion}</td></tr>
            <tr><th scope="row">Active published version</th><td>{draft.active_publication ?? "—"}</td></tr>
            <tr><th scope="row">Published method</th><td>{draft.published_versions?.[String(draft.active_publication)]?.method ?? "—"}</td></tr>
            <tr><th scope="row">Configured visibility</th><td>{draft.public ? "Public" : "Private"}</td></tr>
            <tr><th scope="row">Route</th><td><code>{route}</code> <button type="button" className="secondary" onClick={() => { if (!navigator.clipboard) { report("Clipboard unavailable", true); return; } void navigator.clipboard.writeText(route).then(() => report("URL copied"), () => report("Could not copy URL", true)); }}>Copy URL</button></td></tr>
          </tbody></table></div> : <div className="empty-panel">No publication recorded. Publish from the editor when ready.</div>}
        </section>
      )}
      {(tab === "analytics" || tab === "logs") && !logError && !logsLoaded && <AdminProgress label="Loading execution logs…" />}
      {tab === "analytics" && !logError && logsLoaded && (
        <div className="metric-grid">
          <article>
            <strong>{logs.length}</strong>
            <span>Requests</span>
          </article>
          <article>
            <strong>
              {logs.length ? `${Math.round((successful / logs.length) * 100)}%` : "—"}
            </strong>
            <span>Success rate</span>
          </article>
          <article>
            <strong>{average === null ? "—" : `${average} ms`}</strong>
            <span>Average latency</span>
          </article>
          <article><strong>{p95 === null ? "—" : `${p95} ms`}</strong><span>p95 latency (loaded logs)</span></article>
        </div>
      )}
      {tab === "logs" && !logError && logsLoaded && (
        <section className="function-log-page">
          <div className="log-summary-grid">
            <article>
              <span>Total executions</span>
              <strong>{logs.length}</strong>
            </article>
            <article>
              <span>Successful</span>
              <strong className="log-success">{successful}</strong>
            </article>
            <article>
              <span>Failed</span>
              <strong className="log-failure">{logs.length - successful}</strong>
            </article>
            <article>
              <span>Average latency</span>
              <strong>{average === null ? "—" : `${average} ms`}</strong>
            </article>
          </div>
          <div className="lifecycle-actions"><label>Search execution logs<input type="search" value={logQuery} onChange={event => { setLogQuery(event.target.value); setLogPage(0); }} /></label><label>Execution outcome<select value={logOutcome} onChange={event => { setLogOutcome(event.target.value as LogOutcome); setLogPage(0); }}><option value="all">All outcomes</option><option value="success">Success</option><option value="failure">Failed</option></select></label><button type="button" className="secondary" onClick={() => setLogRefresh(value => value + 1)}>Refresh logs</button></div>
          {!!logs.length && !orderedLogs.length && <p className="muted">No execution logs match the filters.</p>}
          {!logs.length && (
            <div className="empty-panel">
              <strong>No executions yet</strong>
              <p className="muted">
                Every invocation will appear here with its lifecycle status, timing, reason, and
                error details.
              </p>
            </div>
          )}
          <div className="table-wrap compact-log-table">
            <table>
              <thead>
                <tr>
                  <th>Time</th>
                  <th>Status</th>
                  <th>Reason</th>
                  <th>Duration</th>
                  <th>Method</th>
                  <th>Request ID</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {visibleLogs.map((log, index) => {
                  const failed = log.status.toLowerCase() !== "success";
                  return (
                    <tr key={`${log.created_at}-${currentLogPage}-${index}`}>
                      <td><time dateTime={log.created_at}>{new Date(log.created_at).toLocaleString()}</time></td>
                      <td>
                        <span className={`log-status ${failed ? "failure" : "success"}`}>
                          {log.status}
                        </span>
                      </td>
                      <td>
                        {log.reason ?? (failed ? "Execution failed" : "Completed successfully")}
                      </td>
                      <td>{log.duration_ms} ms</td>
                      <td>{log.method ?? "—"}</td>
                      <td>
                        <code>{log.request_id ?? "—"}</code>
                      </td>
                      <td>
                        <button
                          type="button"
                          className="log-view-action"
                          onClick={() => setSelectedLog(log)}
                        >
                          View
                        </button>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
          <div className="log-timeline">
            {selectedLog ? (
              <AdminDialog title="Function execution details" scope={project + " · " + name} onClose={() => setSelectedLog(null)}>
                  {" "}
                  {[selectedLog].map((log, index) => {
                    const failed = log.status.toLowerCase() !== "success";
                    const reason =
                      log.reason ?? (failed ? "Execution failed" : "Completed successfully");
                    return (
                      <article
                        className={`log-entry ${failed ? "failed" : "succeeded"}`}
                        key={`${log.created_at}-${logPage}-${index}`}
                      >
                        <div className="log-entry-marker">{failed ? "!" : "✓"}</div>
                        <div className="log-entry-content">
                          <div className="log-entry-heading">
                            <div>
                              <strong>
                                {failed ? "Failed execution" : "Successful execution"}
                              </strong>
                              <span className={`log-status ${failed ? "failure" : "success"}`}>
                                {log.status}
                              </span>
                            </div>
                            <time dateTime={log.created_at}>
                              {new Date(log.created_at).toLocaleString()}
                            </time>
                            <button
                              type="button"
                              className="log-view-action"
                              onClick={() => setSelectedLog(null)}
                            >
                              Close details
                            </button>
                          </div>
                          <div className="log-detail-grid">
                            <div>
                              <small>Reason</small>
                              <strong>{reason}</strong>
                            </div>
                            <div>
                              <small>Duration</small>
                              <strong>{log.duration_ms} ms</strong>
                            </div>
                            <div>
                              <small>Method</small>
                              <strong>{log.method ?? "—"}</strong>
                            </div>
                            <div>
                              <small>Request ID</small>
                              <strong>{log.request_id ?? "Not recorded"}</strong>
                            </div>
                          </div>
                          <div className="log-latency-breakdown">
                            <small>Latency breakdown</small>
                            {log.steps?.length ? (
                              log.steps.map((step) => (
                                <div
                                  className="latency-row"
                                  key={`latency-${step.name}-${step.duration_ms}`}
                                >
                                  <span>{step.name}</span>
                                  <div>
                                    <i
                                      style={{
                                        width: `${Math.min(100, log.duration_ms ? Math.max(2, (step.duration_ms / log.duration_ms) * 100) : 2)}%`,
                                      }}
                                    />
                                  </div>
                                  <strong>{step.duration_ms} ms</strong>
                                  <em>
                                    {log.duration_ms
                                      ? `${Math.round((step.duration_ms / log.duration_ms) * 100)}%`
                                      : "—"}
                                  </em>
                                </div>
                              ))
                            ) : (
                              <div className="latency-row legacy-latency-row">
                                <span>legacy execution</span>
                                <div>
                                  <i style={{ width: "100%" }} />
                                </div>
                                <strong>{log.duration_ms} ms</strong>
                                <em>100%</em>
                              </div>
                            )}
                            <div className="latency-total">
                              <span>Total workflow latency</span>
                              <strong>{log.duration_ms} ms</strong>
                            </div>
                          </div>
                          <div className={`log-message ${failed ? "error" : "info"}`}>
                            <small>{failed ? "Error details" : "Execution details"}</small>
                            <code>
                              {log.error ??
                                log.message ??
                                (failed
                                  ? "No error details recorded."
                                  : "Function completed without a message.")}
                            </code>
                          </div>
                          <details className="log-raw-record" open>
                            <summary>Raw execution record</summary>
                            <pre>{JSON.stringify(log, null, 2)}</pre>
                          </details>
                          {log.steps?.length ? (
                            <div className="log-steps">
                              <small>Workflow steps and latency</small>
                              {log.steps.map((step) => (
                                <div className="log-step" key={`${step.name}-${step.duration_ms}`}>
                                  <span
                                    className={`log-step-status ${step.status === "success" ? "success" : "failure"}`}
                                  >
                                    {step.status === "success" ? "✓" : "!"}
                                  </span>
                                  <strong>{step.name}</strong>
                                  <span>{step.duration_ms} ms</span>
                                  <small>{step.reason ?? step.error ?? "—"}</small>
                                </div>
                              ))}
                            </div>
                          ) : null}
                        </div>
                      </article>
                    );
                  })}
              </AdminDialog>
            ) : null}
          </div>
          {orderedLogs.length > 10 && (
            <div className="log-pagination">
              <button
                type="button"
                disabled={currentLogPage === 0}
                onClick={() => setLogPage(currentLogPage - 1)}
              >
                Newer
              </button>
              <span>
                Page {currentLogPage + 1} of {logPageCount} · Showing {visibleLogs.length} of{" "}
                {orderedLogs.length}
              </span>
              <button
                type="button"
                disabled={currentLogPage >= logPageCount - 1}
                onClick={() => setLogPage(currentLogPage + 1)}
              >
                Older
              </button>
            </div>
          )}
        </section>
      )}
    </>
  );
}
