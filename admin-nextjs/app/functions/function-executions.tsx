"use client";
import { useEffect, useState } from "react";
import { adminHeaders, projectAdminURL, readAdminSession } from "../lib/admin-api";
import { AdminDialog, AdminProgress } from "../components/admin-dialog";
import { filterFunctionLogs, functionLogMetrics, type FunctionLog, type LogOutcome } from "../lib/function-inspection";
import { formatFunctionDate, type FunctionScope } from "../lib/function-studio";

export function FunctionExecutions({ scope, name, revision, analytics }: { scope: FunctionScope; name: string; revision: number; analytics: boolean }) {
  const [logs, setLogs] = useState<FunctionLog[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [query, setQuery] = useState("");
  const [outcome, setOutcome] = useState<LogOutcome>("all");
  const [page, setPage] = useState(0);
  const [refresh, setRefresh] = useState(0);
  const [live, setLive] = useState(true);
  const [selected, setSelected] = useState<FunctionLog | null>(null);
  useEffect(() => {
    const controller = new AbortController();
    let timer: ReturnType<typeof setTimeout>;
    setLoading(true); setError(""); setSelected(null); setLogs([]);
    async function poll() {
      try {
        const response = await fetch(projectAdminURL(scope.project, "function-logs") + "?name=" + encodeURIComponent(name), { headers: adminHeaders(scope.token, { "x-brain-quiet": "1" }), signal: controller.signal });
        if (!response.ok) throw new Error("Execution history unavailable (" + response.status + ").");
        const rows = (await response.json() as (FunctionLog & { project_id?: string; name?: string })[] | null) ?? [];
        if (!controller.signal.aborted && readAdminSession() === scope.token) {
          setLogs(rows.filter(row => row.project_id === scope.project && row.name === name));
          setError("");
        }
      } catch (reason) {
        if (!controller.signal.aborted && readAdminSession() === scope.token) { setLogs([]); setSelected(null); setError(String(reason)); }
      } finally {
        if (!controller.signal.aborted && readAdminSession() === scope.token) { setLoading(false); if (live) timer = setTimeout(() => void poll(), 5000); }
      }
    }
    void poll();
    return () => { controller.abort(); clearTimeout(timer); };
  }, [scope, name, refresh, live, revision]);
  const { successful, average, p95 } = functionLogMetrics(logs);
  const filtered = filterFunctionLogs(logs, query, outcome);
  const pages = Math.max(1, Math.ceil(filtered.length / 15));
  const currentPage = Math.min(page, pages - 1);
  const total = logs.length;
  return <section className="fn-workspace">
    <div className="fn-heading"><div><h2>{analytics ? "Execution analytics" : "Execution history"}</h2><p className="muted">Metrics reflect loaded host records, including draft tests and published calls.</p></div><div className="fn-actions"><label className="checkbox-label"><input type="checkbox" checked={live} onChange={event => setLive(event.target.checked)} /> Live refresh</label><button type="button" className="secondary" onClick={() => setRefresh(value => value + 1)}>Refresh logs</button></div></div>
    {error ? <p role="alert" className="fn-error">{error}</p> : loading ? <AdminProgress label="Loading executions…" /> : <>
      <div className="fn-metrics">{[["Executions", total], ["Success rate", total ? Math.round(successful / total * 100) + "%" : "—"], ["Average latency", average === null ? "—" : average + " ms"], ["p95 latency", p95 === null ? "—" : p95 + " ms"]].map(([label, value]) => <article key={label}><span>{label}</span><strong>{value}</strong></article>)}</div>
      {analytics ? <section className="fn-panel"><h3>Execution outcomes</h3>{total ? <><div className="fn-outcome-bar" role="img" aria-label={successful + " successful, " + (total - successful) + " failed"}><span style={{ width: successful / total * 100 + "%" }} /></div><div className="fn-actions"><span>{successful} successful</span><span>{total - successful} failed</span></div></> : <p className="muted">No execution records available.</p>}<p className="muted">This is not a time-series telemetry service. Counts and percentile are based only on returned records.</p></section> :
        <section className="fn-panel"><div className="fn-toolbar"><label className="fn-search">Search execution logs<input type="search" value={query} onChange={event => { setQuery(event.target.value); setPage(0); }} /></label><label>Execution outcome<select value={outcome} onChange={event => { setOutcome(event.target.value as LogOutcome); setPage(0); }}><option value="all">All outcomes</option><option value="success">Success</option><option value="failure">Failed</option></select></label></div>
        {!filtered.length ? <div className="fn-empty"><strong>{total ? "No execution logs match the filters." : "No executions yet."}</strong><p className="muted">{total ? "Change your search or outcome filter." : "Run a test or call the published endpoint to generate records."}</p></div> : <div className="table-wrap"><table className="fn-table"><thead><tr><th>Time</th><th>Status</th><th>Duration</th><th>Request ID</th><th>Message</th><th>Details</th></tr></thead><tbody>{filtered.slice(currentPage * 15, currentPage * 15 + 15).map((log, index) => <tr key={log.request_id || log.created_at + index}><td>{formatFunctionDate(log.created_at)}</td><td><span className={"fn-tag " + (log.status.toLowerCase() === "success" ? "fn-tag-live" : "fn-tag-error")}>{log.status}</span></td><td>{log.duration_ms} ms</td><td><code>{log.request_id || "—"}</code></td><td className="fn-log-message">{log.reason || log.message || log.error || "—"}</td><td><button className="secondary" type="button" onClick={() => setSelected(log)}>View</button></td></tr>)}</tbody></table></div>}
        {pages > 1 && <div className="fn-pagination"><button className="secondary" type="button" disabled={!currentPage} onClick={() => setPage(currentPage - 1)}>Newer</button><span>Page {currentPage + 1} of {pages}</span><button className="secondary" type="button" disabled={currentPage === pages - 1} onClick={() => setPage(currentPage + 1)}>Older</button></div>}</section>}
      {selected && <AdminDialog title="Function execution details" scope={scope.project + " · " + name} onClose={() => setSelected(null)}><div className="fn-facts"><div><span>Status</span><strong>{selected.status}</strong></div><div><span>Duration</span><strong>{selected.duration_ms} ms</strong></div><div><span>Request ID</span><code>{selected.request_id || "—"}</code></div></div><p>{selected.error || selected.message || selected.reason || "No execution message recorded."}</p><h3>Workflow steps</h3>{selected.steps?.length ? selected.steps.map((step, index) => <div className="fn-step" key={step.name + index}><strong>{step.name}</strong><span>{step.status} · {step.duration_ms} ms</span><small>{step.reason || step.error || "—"}</small></div>) : <p className="muted">No step timings recorded.</p>}<details><summary>Raw execution record</summary><pre className="fn-output">{JSON.stringify(selected, null, 2)}</pre></details></AdminDialog>}
    </>}
  </section>;
}
