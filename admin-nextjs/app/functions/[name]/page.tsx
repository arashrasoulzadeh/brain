"use client";
import { adminHeaders, readAdminSession } from "../../lib/admin-api";
import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import { AdminPageFrame } from "../../components/admin-chrome";
import { FunctionEditor } from "../function-editor";
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
type Log = {
  status: string;
  duration_ms: number;
  created_at: string;
  message?: string;
  reason?: string;
  error?: string;
  method?: string;
  request_id?: string;
  steps?: { name: string; status: string; duration_ms: number; reason?: string; error?: string }[];
};
type Deployment = { version: number; method: string; route: string; published_at: string };
export default function FunctionDetailPage() {
  const params = useParams<{ name: string }>();
  const router = useRouter();
  const name = decodeURIComponent(params.name ?? "");
  const [pluginAccess, setPluginAccess] = useState("checking");
  const [session, setSession] = useState("");
  const [project, setProject] = useState("");
  const [draft, setDraft] = useState<Draft | null>(null);
  const [logs, setLogs] = useState<Log[]>([]);
  const [deployments, setDeployments] = useState<Deployment[]>([]);
  const [tab, setTab] = useState("editor");
  const [message, setMessage] = useState("");
  const [route, setRoute] = useState("");
  const [publishing, setPublishing] = useState(false);
  const [validating, setValidating] = useState(false);
  const [validationMessage, setValidationMessage] = useState("");
  const [logPage, setLogPage] = useState(0);
  const [selectedLog, setSelectedLog] = useState<Log | null>(null);
  async function refresh(token: string) {
    const headers = adminHeaders(token);
    const projectsResponse = await fetch("/api/admin/projects", { headers });
    if (!projectsResponse.ok)
      throw new Error((await projectsResponse.text()) || "Could not load projects.");
    const projects = (await projectsResponse.json()) as { ID: string; Active: boolean }[];
    const active = projects.find((candidate) => candidate.Active)?.ID ?? "";
    setProject(active);
    if (!active) throw new Error("No active project selected.");
    const draftsResponse = await fetch(
      `/api/admin/functions?project_id=${encodeURIComponent(active)}`,
      { headers },
    );
    if (!draftsResponse.ok)
      throw new Error((await draftsResponse.text()) || "Could not load function.");
    const drafts = (await draftsResponse.json()) as Draft[];
    const current = drafts.find((candidate) => candidate.name === name) ?? null;
    setDraft(current);
    const logsResponse = await fetch(
      `/api/admin/function-logs?project_id=${encodeURIComponent(active)}&name=${encodeURIComponent(name)}`,
      { headers },
    );
    if (logsResponse.ok) setLogs((await logsResponse.json()) as Log[]);
    if (current?.published_at)
      setDeployments([
        {
          version: current.version ?? 1,
          method: current.method ?? "POST",
          route: `${window.location.origin}/public/${active}/${name}`,
          published_at: current.published_at,
        },
      ]);
  }
  async function refreshLogs(token: string, projectID: string) {
    const response = await fetch(
      `/api/admin/function-logs?project_id=${encodeURIComponent(projectID)}&name=${encodeURIComponent(name)}`,
      { headers: adminHeaders(token) },
    );
    if (response.ok) setLogs((await response.json()) as Log[]);
  }
  useEffect(() => {
    const token = readAdminSession();
    setSession(token);
    if (!token) {
      router.replace("/");
      return;
    }
    void fetch("/api/admin/projects", { headers: adminHeaders(token) })
      .then(async (response) => {
        if (!response.ok) throw new Error("session expired");
        const projects = (await response.json()) as { ID: string; Active: boolean }[];
        const active = projects.find((candidate) => candidate.Active)?.ID ?? "";
        if (!active) throw new Error("no active project");
        const pluginsResponse = await fetch(
          `/api/admin/plugins?project_id=${encodeURIComponent(active)}`,
          { headers: adminHeaders(token) },
        );
        if (!pluginsResponse.ok) throw new Error("plugin access unavailable");
        const payload = (await pluginsResponse.json()) as {
          installations: { PluginID: string; Enabled: boolean; Ready: boolean }[];
        };
        if (
          !payload.installations.some(
            (installation) =>
              installation.PluginID === "functions" && installation.Enabled && installation.Ready,
          )
        ) {
          router.replace("/");
          return;
        }
        setPluginAccess("allowed");
        void refresh(token);
      })
      .catch(() => router.replace("/"));
  }, [name, router]);
  useEffect(() => {
    if (pluginAccess !== "allowed" || !session || !project) return;
    const refreshTimer = window.setInterval(() => void refreshLogs(session, project), 1000);
    return () => window.clearInterval(refreshTimer);
  }, [pluginAccess, project, session, name]);
  async function savePermissions(permissions: string[], publicRoute: boolean) {
    if (!draft) return;
    const response = await fetch("/api/admin/functions", {
      method: "POST",
      headers: adminHeaders(session, { "content-type": "application/json" }),
      body: JSON.stringify({
        operation: "functions.permissions",
        project_id: project,
        name: draft.name,
        permissions,
        public: publicRoute,
      }),
    });
    setMessage(response.ok ? "Settings saved." : await response.text());
    if (response.ok) await refresh(session);
  }
  async function publish() {
    if (!draft || !project || !session) {
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
      const update = await fetch("/api/admin/functions", {
        method: "POST",
        headers,
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
      const response = await fetch("/api/admin/functions", {
        method: "POST",
        headers,
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
        setMessage(body || `Publish failed (${response.status}).`);
        return;
      }
      setMessage("Published current version.");
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
      await refresh(session);
    } catch (error) {
      setMessage(error instanceof Error ? error.message : "Publish request failed.");
    } finally {
      setPublishing(false);
    }
  }
  async function validateSource() {
    if (!draft || !project || !session) {
      setValidationMessage("Select an active project and sign in before validating.");
      return;
    }
    setValidating(true);
    setValidationMessage("");
    try {
      const response = await fetch("/api/admin/functions", {
        method: "POST",
        headers: adminHeaders(session, { "content-type": "application/json" }),
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
      setValidationMessage("Source passed host validation and is ready for sandbox compilation.");
    } catch (error) {
      setValidationMessage(error instanceof Error ? error.message : "Validation request failed.");
    } finally {
      setValidating(false);
    }
  }
  async function saveDraft() {
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
      const response = await fetch("/api/admin/functions", {
        method: "POST",
        headers: adminHeaders(session, { "content-type": "application/json" }),
        body: JSON.stringify({
          operation: "functions.update",
          project_id: project,
          name: draft.name,
          source: draft.source,
          function_version: draft.version ?? 1,
        }),
      });
      setMessage(response.ok ? "Changes saved." : await response.text());
      if (response.ok) await refresh(session);
    } finally {
      setPublishing(false);
    }
  }
  if (pluginAccess !== "allowed") return <p className="muted">Checking page access…</p>;
  if (!draft)
    return (
      <AdminPageFrame
        activePage="functions"
        title={name}
        projectID={project}
        requiredPluginID="functions"
      >
        <p className="muted">Function not found or still loading.</p>
      </AdminPageFrame>
    );
  const permissions = draft.permissions ?? ["functions.invoke"];
  const successful = logs.filter((log) => log.status === "success").length;
  const average = logs.length
    ? Math.round(logs.reduce((sum, log) => sum + log.duration_ms, 0) / logs.length)
    : 0;
  const orderedLogs = logs.slice().reverse();
  const visibleLogs = orderedLogs.slice(logPage * 10, logPage * 10 + 10);
  const logPageCount = Math.max(1, Math.ceil(orderedLogs.length / 10));
  return (
    <AdminPageFrame
      activePage="functions"
      title={draft.name}
      projectID={project}
      requiredPluginID="functions"
    >
      <div className="section-heading">
        <Link href="/functions">Back to functions</Link>
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
      {tab === "editor" && (
        <>
          <div className="setting-toolbar flex flex-wrap items-end gap-4 rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
            <div className="rounded-lg bg-slate-100 px-3 py-2 text-sm font-semibold text-slate-700">
              Version {draft.version ?? 1}
            </div>
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
              disabled={publishing}
              className="h-10 rounded-lg bg-slate-700 px-4 text-sm font-semibold text-white shadow-sm transition hover:bg-slate-800 disabled:cursor-not-allowed disabled:opacity-60"
              onClick={() => void saveDraft()}
            >
              {publishing ? "Saving…" : "Save changes"}
            </button>
            <button
              type="button"
              disabled={publishing}
              className="h-10 rounded-lg bg-indigo-600 px-4 text-sm font-semibold text-white shadow-sm transition hover:bg-indigo-700 disabled:cursor-not-allowed disabled:opacity-60"
              onClick={() => void publish()}
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
              Public route: <code>{route}</code>
            </p>
          )}
          <FunctionEditor
            source={draft.source}
            onChange={(source) => setDraft({ ...draft, source })}
            onValidate={() => void validateSource()}
            validating={validating}
            validationMessage={validationMessage}
          />
          <div className="permission-panel">
            <h2>Function settings</h2>
            {permissions.map((permission) => (
              <div className="permission-row" key={permission}>
                <code>{permission}</code>
                <button
                  className="danger"
                  onClick={() =>
                    void savePermissions(
                      permissions.filter((candidate) => candidate !== permission),
                      Boolean(draft.public),
                    )
                  }
                >
                  Remove
                </button>
              </div>
            ))}
            <button
              onClick={() =>
                void savePermissions([...permissions, "functions.invoke"], Boolean(draft.public))
              }
            >
              Add invoke permission
            </button>
            <label className="checkbox-label">
              <input
                type="checkbox"
                checked={Boolean(draft.public)}
                onChange={(event) => void savePermissions(permissions, event.target.checked)}
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
                <th>Public URL</th>
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
                          `${window.location.origin}${deployment.route}`,
                        )
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
            <strong>{logs.length ? `${average} ms` : "—"}</strong>
            <span>Average latency</span>
          </article>
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
              <strong>{logs.length ? `${average} ms` : "—"}</strong>
            </article>
          </div>
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
                    <tr key={`${log.created_at}-${logPage}-${index}`}>
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
              <div
                className="log-modal-backdrop"
                role="presentation"
                onClick={() => setSelectedLog(null)}
              >
                <div
                  className="log-modal"
                  role="dialog"
                  aria-modal="true"
                  aria-label="Function execution details"
                  onClick={(event) => event.stopPropagation()}
                >
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
                </div>
              </div>
            ) : null}
          </div>
          {orderedLogs.length > 10 && (
            <div className="log-pagination">
              <button
                type="button"
                disabled={logPage === 0}
                onClick={() => setLogPage((page) => Math.max(0, page - 1))}
              >
                Newer
              </button>
              <span>
                Page {logPage + 1} of {logPageCount} · Showing {visibleLogs.length} of{" "}
                {orderedLogs.length}
              </span>
              <button
                type="button"
                disabled={logPage >= logPageCount - 1}
                onClick={() => setLogPage((page) => Math.min(logPageCount - 1, page + 1))}
              >
                Older
              </button>
            </div>
          )}
        </section>
      )}
    </AdminPageFrame>
  );
}
