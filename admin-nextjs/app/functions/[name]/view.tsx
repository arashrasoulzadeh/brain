"use client";
import { adminHeaders, readAdminSession, preferredAdminProject, projectAdminURL, systemAdminURL, projectUIURL, systemUIURL, adminUIContext } from "../../lib/admin-api";
import { useEffect, useRef, useState } from "react";
import { useParams, usePathname } from "next/navigation";
import Link from "next/link";
import { AdminDialog, AdminProgress } from "../../components/admin-dialog";
import { AdminPageFrame } from "../../components/admin-chrome";
import { FunctionEditor } from "../function-editor";
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
};
type Deployment = { version: number; method: string; route: string; published_at: string };
export default function FunctionDetailPage() {
  const params = useParams<{ name: string }>();
  const { namespace: requestedProject } = adminUIContext(usePathname());
  const name = decodeURIComponent(params.name ?? "");
  return <AdminPageFrame activePage="functions" title={name} requiredPluginID="functions"><FunctionDetail key={(requestedProject ?? "") + ":" + name} name={name} requestedProject={requestedProject} /></AdminPageFrame>;
}
function FunctionDetail({ name, requestedProject }: { name: string; requestedProject?: string }) {
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState("");
  const [logError, setLogError] = useState("");
  const [validationOK, setValidationOK] = useState(false);
  const [settingsPending, setSettingsPending] = useState<{ permissions: string[]; publicRoute: boolean }>();
  const [publishPending, setPublishPending] = useState(false);
  const [session, setSession] = useState("");
  const [project, setProject] = useState("");
  const [draft, setDraft] = useState<Draft | null>(null);
  const [savedSource, setSavedSource] = useState("");
  const [discardPending, setDiscardPending] = useState(false);
  const dirty = draft !== null && draft.source !== savedSource;
  useEffect(() => {
    if (!dirty) return;
    function warn(event: BeforeUnloadEvent) { event.preventDefault(); event.returnValue = ""; }
    window.addEventListener("beforeunload", warn);
    return () => window.removeEventListener("beforeunload", warn);
  }, [dirty]);
  const [logs, setLogs] = useState<Log[]>([]);
  const [deployments, setDeployments] = useState<Deployment[]>([]);
  const [tab, setTab] = useState("editor");
  const [message, setMessage] = useState("");
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
    setDraft(current && preserveSource !== undefined ? { ...current, source: preserveSource } : current);
    const currentRoute = current?.published_at ? window.location.origin + "/" + encodeURIComponent(selected?.Slug || active) + "/functions/" + encodeURIComponent(name) : "";
    setRoute(currentRoute);
    setDeployments(current?.published_at ? [{ version: current.version ?? 1, method: current.method ?? "POST", route: currentRoute, published_at: current.published_at }] : []);
  }
  useEffect(() => {
    const controller = new AbortController();
    lifetime.current = controller;
    const token = readAdminSession();
    setSession(token); setLoading(true); setLoadError("");
    if (!token) { setLoadError("Sign in to continue"); setLoading(false); return () => controller.abort(); }
    void refresh(token).catch(reason => { if (!controller.signal.aborted) setLoadError(String(reason)); }).finally(() => { if (!controller.signal.aborted) setLoading(false); });
    return () => { controller.abort(); if (lifetime.current === controller) lifetime.current = null; };
  }, [name, requestedProject]);
  useEffect(() => {
    if (!session || !project) return;
    const controller = new AbortController();
    let timer: ReturnType<typeof setTimeout>;
    async function poll() {
      try {
        const response = await fetch(projectAdminURL(project, "function-logs") + "?name=" + encodeURIComponent(name), { headers: adminHeaders(session, { "x-brain-quiet": "1" }), signal: controller.signal });
        if (!response.ok) throw new Error("Function logs unavailable");
        const rows = (await response.json() as Log[] | null) ?? [];
        if (!controller.signal.aborted && readAdminSession() === session) { setLogs(rows); setLogError(""); setLogPage(page => Math.min(page, Math.max(0, Math.ceil(rows.length / 10) - 1))); }
      } catch (reason) { if (!controller.signal.aborted) { setLogs([]); setSelectedLog(null); setLogError(String(reason)); } }
      finally { if (!controller.signal.aborted) timer = setTimeout(() => void poll(), 5000); }
    }
    void poll();
    return () => { controller.abort(); clearTimeout(timer); };
  }, [project, session, name, logRefresh]);
  async function savePermissions(permissions: string[], publicRoute: boolean) {
    if (!draft || publishing) return;
    setPublishing(true); setMessage("");
    try {
      const response = await fetch(projectAdminURL(project, "functions"), {
        method: "POST", headers: adminHeaders(session, { "content-type": "application/json" }), signal: lifetime.current?.signal,
        body: JSON.stringify({ operation: "functions.permissions", project_id: project, name: draft.name, permissions, public: publicRoute }),
      });
      if (!response.ok) throw new Error(await response.text());
      setDraft(previous => previous ? { ...previous, permissions, public: publicRoute } : previous);
      setMessage("Settings saved."); setSettingsPending(undefined);
    } catch (reason) { setMessage(String(reason)); }
    finally { setPublishing(false); }
  }
  async function publish() {
    if (publishing || !draft || !project || !session) {
      setMessage("Select an active project and sign in before publishing.");
      return;
    }
    if (!draft.name.trim() || !draft.source.trim()) {
      setMessage("Cannot publish: function name and source are required.");
      return;
    }
    setPublishing(true);
    setMessage("Saving source and publishing…");
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
          function_version: draft.version ?? 1,
        }),
      });
      if (!update.ok) {
        setMessage((await update.text()) || `Source save failed (${update.status}).`);
        return;
      }
      setSavedSource(draft.source);
      const response = await fetch(projectAdminURL(project, "functions"), {
        method: "POST",
        headers, signal: lifetime.current?.signal,
        body: JSON.stringify({
          operation: "functions.publish",
          project_id: project,
          name: draft.name,
          function_version: draft.version ?? 1,
          permissions: draft.permissions ?? ["functions.invoke"],
        }),
      });
      const body = await response.text();
      let result: {
        route?: string;
        version?: number;
        method?: string;
        published_at?: string;
      } | null = null;
      try {
        result = JSON.parse(body);
      } catch {}
      if (!response.ok) {
        setMessage("Source saved; publication failed: " + (body || String(response.status)));
        return;
      }
      setMessage("Published current version.");
      setPublishPending(false);
      if (result?.route) {
        const publicRoute = `${window.location.origin}${new URL(result.route, window.location.origin).pathname}`;
        setRoute(publicRoute);
        setDeployments((items) => [
          {
            version: result.version ?? draft.version ?? 1,
            method: result.method ?? draft.method ?? "POST",
            route: publicRoute,
            published_at: result.published_at ?? new Date().toISOString(),
          },
          ...items,
        ]);
      }
      await refresh(session).catch(reason => setMessage("Published; refresh failed: " + String(reason)));
    } catch (error) {
      setMessage(error instanceof Error ? error.message : "Publish request failed.");
    } finally {
      setPublishing(false);
    }
  }
  async function validateSource() {
    if (validating || publishing) return;
    if (!draft || !project || !session) {
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
    if (!draft || !project || !session) {
      setMessage("Select an active project and sign in before saving.");
      return;
    }
    if (!draft.name.trim() || !draft.source.trim()) {
      setMessage("Cannot save: function name and source are required.");
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
          function_version: draft.version ?? 1,
        }),
      });
      setMessage(response.ok ? "Changes saved." : await response.text());
      if (response.ok) { setSavedSource(draft.source); await refresh(session).catch(reason => setMessage("Source saved; refresh failed: " + String(reason))); }
    } catch (reason) { setMessage(String(reason)); } finally {
      setPublishing(false);
    }
  }
  if (loading) return <AdminProgress label="Loading function…" />;
  if (loadError) return <p role="alert">{loadError}</p>;
  if (!draft) return <><p role="status">Function not found.</p>{project && <Link href={projectUIURL(project, "functions")}>Back to functions</Link>}</>;
  const permissions = draft.permissions ?? ["functions.invoke"];
  const { successful, average, p95 } = functionLogMetrics(logs);
  const orderedLogs = filterFunctionLogs(logs, logQuery, logOutcome);
  const logPageCount = Math.max(1, Math.ceil(orderedLogs.length / 10));
  const currentLogPage = Math.min(logPage, logPageCount - 1);
  const visibleLogs = orderedLogs.slice(currentLogPage * 10, currentLogPage * 10 + 10);
  return (
    <>
      <div className="section-heading">
        <Link href={project ? projectUIURL(project, "functions") : systemUIURL("projects")}>Back to functions</Link>
      </div>
      <div className="function-tabs">
        {[
          ["editor", "Editor"],
          ["deployments", "Deployments"],
          ["analytics", "Analytics"],
          ["logs", "Logs"],
        ].map(([value, label]) => (
          <button
            type="button"
            className={`${tab === value ? "active" : ""} rounded-lg px-4 py-2 text-sm font-semibold transition-colors`}
            onClick={() => setTab(value)}
            key={value}
          >
            {label}
          </button>
        ))}
      </div>
      {discardPending && <AdminDialog title="Discard unsaved source" scope={project + " · " + name} onClose={() => setDiscardPending(false)}><p>Restore the last saved source? Local edits will be lost.</p><button type="button" onClick={() => { setDraft(previous => previous ? { ...previous, source: savedSource } : previous); setValidationMessage(""); setValidationOK(false); setDiscardPending(false); }}>Discard edits</button></AdminDialog>}
      {publishPending && <AdminDialog title="Publish current version" scope={project + " · " + name} busy={publishing} onClose={() => setPublishPending(false)}><p>Save the current source and publish version {draft.version ?? 1}?</p><p>{draft.public ? "This route is configured for public access." : "This route requires authorization."}</p><button type="button" disabled={publishing} onClick={() => void publish()}>Confirm publish</button>{publishing && <AdminProgress label="Saving and publishing…" />}{message && <p role="status">{message}</p>}</AdminDialog>}
      {settingsPending && <AdminDialog title="Update function settings" scope={project + " · " + name} busy={publishing} onClose={() => setSettingsPending(undefined)}><p>{settingsPending.publicRoute ? "Public route enabled" : "Private route"} · Permissions: {settingsPending.permissions.join(", ") || "None"}</p><button type="button" disabled={publishing} onClick={() => void savePermissions(settingsPending.permissions, settingsPending.publicRoute)}>Confirm settings</button>{publishing && <AdminProgress label="Saving function settings…" />}{message && <p role="status">{message}</p>}</AdminDialog>}
      {tab !== "editor" && message && <p role="status">{message}</p>}
      {logError && (tab === "logs" || tab === "analytics") && <p role="alert">{logError}</p>}
      {tab === "editor" && (
        <>
          <div className="setting-toolbar flex flex-wrap items-end gap-4 rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
            <div className="rounded-lg bg-slate-100 px-3 py-2 text-sm font-semibold text-slate-700">
              Version {draft.version ?? 1}
            </div>
            <span role="status">{dirty ? "Unsaved source changes" : "Source saved"}</span>
            <button type="button" className="secondary" disabled={!dirty || publishing || validating} onClick={() => setDiscardPending(true)}>Restore saved source</button>
            <div
              className={`rounded-lg px-3 py-2 text-sm font-semibold ${draft.public ? "bg-emerald-100 text-emerald-800" : "bg-slate-100 text-slate-700"}`}
            >
              {draft.public ? "Public" : "Private"}
            </div>
            <div className="rounded-lg bg-slate-100 px-3 py-2 text-sm font-semibold text-slate-700">
              Method {draft.method ?? "—"} · set in source config
            </div>
            <button
              type="button"
              disabled={publishing || validating}
              className="h-10 rounded-lg bg-slate-700 px-4 text-sm font-semibold text-white shadow-sm transition hover:bg-slate-800 disabled:cursor-not-allowed disabled:opacity-60"
              onClick={() => void saveDraft()}
            >
              {publishing ? "Saving…" : "Save changes"}
            </button>
            <button
              type="button"
              disabled={publishing || validating}
              className="h-10 rounded-lg bg-indigo-600 px-4 text-sm font-semibold text-white shadow-sm transition hover:bg-indigo-700 disabled:cursor-not-allowed disabled:opacity-60"
              onClick={() => { setMessage(""); setPublishPending(true); }}
            >
              {publishing ? "Publishing…" : "Publish current version"}
            </button>
          </div>
          {message && (
            <p className="notice" role="status" aria-live="polite">
              {message}
            </p>
          )}
          {route && (
            <p className="notice">
              Published route: <code>{route}</code>
            </p>
          )}
          <FunctionEditor
            source={draft.source}
            onChange={(source) => { setDraft({ ...draft, source }); setValidationMessage(""); setValidationOK(false); }}
            readOnly={publishing || validating}
            onValidate={() => void validateSource()}
            validating={validating}
            validationMessage={validationMessage}
            validationOK={validationOK}
          />
          <div className="permission-panel">
            <h2>Function settings</h2>
            {permissions.map((permission) => (
              <div className="permission-row" key={permission}>
                <code>{permission}</code>
                <button
                  className="danger"
                  disabled={publishing}
                  onClick={() =>
                    setSettingsPending({ permissions: permissions.filter(candidate => candidate !== permission), publicRoute: Boolean(draft.public) })
                  }
                >
                  Remove
                </button>
              </div>
            ))}
            <button disabled={publishing || permissions.includes("functions.invoke")}
              onClick={() =>
                setSettingsPending({ permissions: [...permissions, "functions.invoke"], publicRoute: Boolean(draft.public) })
              }
            >
              Add invoke permission
            </button>
            <label className="checkbox-label">
              <input
                type="checkbox"
                checked={Boolean(draft.public)}
                disabled={publishing}
                onChange={(event) => setSettingsPending({ permissions, publicRoute: event.target.checked })}
              />{" "}
              Public route enabled
            </label>
            <small>{message}</small>
          </div>
        </>
      )}
      {tab === "deployments" && (
        <div className="table-wrap">
          <table>
            <thead>
              <tr>
                <th>Version</th>
                <th>Method</th>
                <th>Published URL</th>
                <th>Published</th>
                <th />
              </tr>
            </thead>
            <tbody>
              {deployments.map((deployment, index) => (
                <tr key={`${deployment.version}-${index}`}>
                  <td>{deployment.version}</td>
                  <td>{deployment.method}</td>
                  <td>
                    <code>{deployment.route}</code>
                  </td>
                  <td>{new Date(deployment.published_at).toLocaleString()}</td>
                  <td>
                    <button
                      onClick={() =>
                        void navigator.clipboard.writeText(
                          new URL(deployment.route, window.location.origin).href,
                        ).catch(() => setMessage("Could not copy URL"))
                      }
                    >
                      Copy URL
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          {!deployments.length && (
            <p className="muted">No deployments yet. Publish the current version to create one.</p>
          )}
        </div>
      )}
      {tab === "analytics" && (
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
      {tab === "logs" && (
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
                      <td>{log.method ?? draft.method ?? "—"}</td>
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
                              <strong>{log.method ?? draft.method ?? "—"}</strong>
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
