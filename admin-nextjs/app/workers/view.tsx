"use client";
import { useEffect, useState } from "react";
import { usePathname, useRouter } from "next/navigation";
import { adminHeaders, readAdminSession, preferredAdminProject, projectAdminURL, systemAdminURL, systemUIURL, adminUIContext, notifyPluginCatalogChanged } from "../lib/admin-api";
import { availableSlots, filterNeurons, legacyWorkerState, summarizeCluster, type ClusterStatus, type Installation, type Neuron, type NeuronAction } from "../lib/neuron-dashboard";
import { AdminDialog, AdminProgress } from "../components/admin-dialog";
import { PluginManageLink } from "../components/plugin-manage-link";
import { AdminPageFrame } from "../components/admin-chrome";

function NeuronRow({ neuron, busy, onAction }: { neuron: Neuron; busy: boolean; onAction: (id: string, action: NeuronAction) => void }) {
  const free = availableSlots(neuron);
  const status = neuron.Draining ? "Draining" : free > 0 ? "Available" : "At capacity";
  const runtime = neuron.Runtimes ?? [];
  return <tr>
    <th scope="row"><strong>{neuron.ID}</strong><small className="table-subtitle muted">Connected via scheduler</small></th>
    <td><span className={"status-pill " + (status === "Available" ? "is-success" : "is-warning")}>{status}</span></td>
    <td><div className="neuron-runtime-list">{runtime.map(item => <span key={item.Kind + ":" + item.Profile}>{item.Kind}<small>{item.Profile}</small></span>)}{!runtime.length && "—"}</div></td>
    <td><div className="neuron-capacity"><progress aria-label={neuron.ID + " occupied slots"} value={Math.min(neuron.Running, neuron.Slots)} max={Math.max(neuron.Slots, 1)} /><span>{neuron.Running} occupied / {neuron.Slots} total · {free} free</span></div></td>
    <td><time dateTime={neuron.LastPulse}>{new Date(neuron.LastPulse).toLocaleString()}</time></td>
    <td>{neuron.OperatorDraining ? <button type="button" className="secondary" disabled={busy} onClick={() => onAction(neuron.ID, "resume")}>Resume {neuron.ID}</button> : <button type="button" className="secondary" disabled={busy || neuron.Draining} onClick={() => onAction(neuron.ID, "drain")}>Drain {neuron.ID}</button>}</td>
  </tr>;
}

export default function WorkersPage() {
  const router = useRouter();
  const { namespace } = adminUIContext(usePathname());
  const [cluster, setCluster] = useState<ClusterStatus>();
  const [legacy, setLegacy] = useState<{ project: string; workers: Installation[] }>();
  const [clusterError, setClusterError] = useState("");
  const [legacyError, setLegacyError] = useState("");
  const [loading, setLoading] = useState(true);
  const [lastRefresh, setLastRefresh] = useState<Date>();
  const [refresh, setRefresh] = useState(0);
  const [search, setSearch] = useState("");
  const [state, setState] = useState("all");
  const [pending, setPending] = useState<Installation>();
  const [neuronAction, setNeuronAction] = useState<{ id: string; action: NeuronAction; token: string }>();
  const [message, setMessage] = useState("");
  const [busy, setBusy] = useState(false);
  const [done, setDone] = useState(false);

  useEffect(() => {
    const controller = new AbortController();
    const token = readAdminSession();
    if (!token) { router.replace(systemUIURL()); return; }
    const valid = () => !controller.signal.aborted && readAdminSession() === token;
    const options = { headers: adminHeaders(token), signal: controller.signal, cache: "no-store" as const };
    let timer: ReturnType<typeof setTimeout>;
    setLoading(true);
    setCluster(undefined);
    setLegacy(undefined);
    setClusterError("");
    setLegacyError("");
    setLastRefresh(undefined);
    async function loadCluster() {
      try {
        const response = await fetch(systemAdminURL("neurons"), options);
        if (!response.ok) throw new Error("Neuron status unavailable");
        const data = await response.json() as ClusterStatus;
        if (valid()) { setCluster(data); setClusterError(""); setLastRefresh(new Date()); }
      } catch (reason) {
        if (valid()) { setCluster(undefined); setLastRefresh(undefined); setClusterError(String(reason)); }
      }
    }
    async function loadLegacy() {
      try {
        const response = await fetch(systemAdminURL("projects"), options);
        if (!response.ok) throw new Error("Projects unavailable");
        const projects = (await response.json() as { ID: string; Slug?: string; Active: boolean }[] | null) ?? [];
        const project = preferredAdminProject(projects, namespace)?.ID;
        if (!project) { if (valid()) { setLegacy(undefined); setLegacyError(""); } return; }
        const plugins = await fetch(projectAdminURL(project, "plugins"), options);
        if (!plugins.ok) throw new Error("Legacy worker status unavailable");
        const data = await plugins.json() as { installations: Installation[] | null };
        if (valid()) { setLegacy({ project, workers: (data.installations ?? []).filter(worker => worker.ProjectID === project) }); setLegacyError(""); }
      } catch (reason) {
        if (valid()) { setLegacy(undefined); setLegacyError(String(reason)); }
      }
    }
    async function poll() {
      if (!valid()) return;
      if (!document.hidden || timer === undefined) await Promise.all([loadCluster(), loadLegacy()]);
      if (valid()) { setLoading(false); timer = setTimeout(() => void poll(), 5000); }
    }
    void poll();
    return () => { controller.abort(); clearTimeout(timer); };
  }, [router, namespace, refresh]);

  async function spinUp() {
    const token = readAdminSession();
    if (busy || done || !pending || pending.ProjectID !== legacy?.project || !token) return;
    setBusy(true); setMessage("");
    try {
      const response = await fetch(projectAdminURL(pending.ProjectID, "actions"), { method: "POST", headers: adminHeaders(token, { "content-type": "application/json" }), body: JSON.stringify({ operation: "plugin.worker.start", project_id: pending.ProjectID, plugin_id: pending.PluginID, version: pending.Version }) });
      if (!response.ok) throw new Error("Worker start failed");
      if (readAdminSession() === token) { setDone(true); notifyPluginCatalogChanged(pending.ProjectID); setRefresh(value => value + 1); }
    } catch (reason) { if (readAdminSession() === token) setMessage(String(reason)); }
    finally { setBusy(false); }
  }

  function requestNeuronAction(id: string, action: NeuronAction) {
    const token = readAdminSession();
    if (!token) return;
    setNeuronAction({ id, action, token }); setDone(false); setMessage("");
  }

  async function applyNeuronAction() {
    const token = readAdminSession();
    if (busy || done || !neuronAction || !token || token !== neuronAction.token) return;
    setBusy(true); setMessage("");
    try {
      const response = await fetch(systemAdminURL("neurons"), { method: "POST", headers: adminHeaders(token, { "content-type": "application/json" }), body: JSON.stringify({ id: neuronAction.id, action: neuronAction.action, confirmed: true }) });
      if (!response.ok) throw new Error("Neuron action failed: " + response.status);
      if (readAdminSession() === token) { setDone(true); setRefresh(value => value + 1); }
    } catch (reason) { if (readAdminSession() === token) { setMessage(String(reason)); setRefresh(value => value + 1); } }
    finally { setBusy(false); }
  }

  const summary = summarizeCluster(cluster);
  const neurons = cluster?.configured ? cluster.snapshot.Neurons ?? [] : [];
  const visible = filterNeurons(neurons, search, state);
  return <AdminPageFrame activePage="workers" title="Neurons" eyebrow="EXECUTION INFRASTRUCTURE" projectID={legacy?.project}>
    <div className="neuron-console">
      <div className="neuron-toolbar"><div><p className="neuron-intro">Manage execution capacity across projects and inspect the Cortex scheduler.</p><p className="muted neuron-updated">{lastRefresh ? "Last updated " + lastRefresh.toLocaleTimeString() + " · Refreshes every 5 seconds" : "No current cluster snapshot"}</p></div><button type="button" className="secondary" disabled={loading || busy} onClick={() => setRefresh(value => value + 1)}>Refresh workers</button></div>
      {loading && <AdminProgress label="Loading execution infrastructure…" />}
      {clusterError && <p role="alert">{clusterError}. No cached metrics are shown.</p>}
      {cluster && !cluster.configured && <section className="neuron-notice"><span className="status-pill is-warning">Not configured</span><h2>Neuron transport is not connected</h2><p>Brain is still using dedicated plugin workers. The Cortex scheduler is not attached to this server, so live Neuron capacity is unavailable.</p></section>}
      {cluster?.configured && summary?.connected === 0 && <section className="neuron-notice"><span className="status-pill is-warning">No capacity</span><h2>No connected Neurons</h2><p>The scheduler is online, but no Neurons have registered.</p></section>}
      <section aria-label="Cortex capacity" className="neuron-metrics metric-grid">
        <article><span>Connected Neurons</span><strong>{summary?.connected ?? "—"}</strong><small className="muted">Authenticated sessions</small></article>
        <article><span>Available slots</span><strong>{summary?.available ?? "—"}{summary && <small> / {summary.slots}</small>}</strong><small className="muted">Excludes draining capacity</small></article>
        <article><span>Running executions</span><strong>{summary?.running ?? "—"}</strong><small className="muted">Active Cortex assignments</small></article>
        <article><span>Pending executions</span><strong>{summary?.queued ?? "—"}</strong><small className="muted">Current queue depth</small></article>
      </section>
      <section className="section" aria-labelledby="neuron-list-title">
        <div className="section-heading"><div><p className="eyebrow">CLUSTER INVENTORY</p><h2 id="neuron-list-title">Connected Neurons</h2></div><span className="count">{summary?.connected ?? "—"}</span></div>
        <div className="neuron-filters"><label>Search<input value={search} onChange={event => setSearch(event.target.value)} placeholder="Identity or runtime" /></label><label>Capacity<select value={state} onChange={event => setState(event.target.value)}><option value="all">All statuses</option><option value="available">Available</option><option value="busy">At capacity</option><option value="draining">Draining</option></select></label></div>
        <div className="table-wrap neuron-table"><table><thead><tr><th>Neuron</th><th>Status</th><th>Runtimes</th><th>Execution slots</th><th>Last heartbeat</th><th>Placement</th></tr></thead><tbody>{visible.map(neuron => <NeuronRow key={neuron.ID} neuron={neuron} busy={busy} onAction={requestNeuronAction} />)}{!visible.length && <tr><td colSpan={6} className="neuron-empty">{!cluster?.configured ? "Neuron inventory unavailable." : neurons.length ? "No Neurons match the filters." : "No Neurons connected."}</td></tr>}</tbody></table></div>
        <p className="muted neuron-note">Runtime usage is launcher process accounting, not an isolation guarantee. Drain stops new assignments; active executions continue. Resume removes the operator pause, not a runtime-initiated drain.</p>
      </section>
      {cluster?.configured && Boolean(cluster.managed?.length) && <section className="section" aria-labelledby="managed-neurons-title">
        <div className="section-heading"><div><p className="eyebrow">LOCAL DEVELOPMENT</p><h2 id="managed-neurons-title">Managed Neuron processes</h2><p className="muted">Stop or restart drains current work before replacing the process. Remote processes are controlled by their own supervisor.</p></div></div>
        <div className="table-wrap neuron-table"><table><thead><tr><th>Identity</th><th>Process</th><th>Connection</th><th>Actions</th></tr></thead><tbody>
          {cluster.managed?.map(process => <tr key={process.ID}><th scope="row">{process.ID}</th><td>{process.Running ? "Running" : "Stopped"}</td><td>{neurons.some(neuron => neuron.ID === process.ID) ? "Connected" : "Disconnected"}</td><td><div className="table-actions">{process.Running ? <><button type="button" className="secondary" disabled={busy} onClick={() => requestNeuronAction(process.ID, "stop")}>Stop {process.ID}</button><button type="button" className="secondary" disabled={busy} onClick={() => requestNeuronAction(process.ID, "restart")}>Restart {process.ID}</button></> : <button type="button" className="secondary" disabled={busy} onClick={() => requestNeuronAction(process.ID, "start")}>Start {process.ID}</button>}</div></td></tr>)}
        </tbody></table></div>
      </section>}
      <section className="section" aria-labelledby="execution-history-title">
        <div className="section-heading"><div><p className="eyebrow">EXECUTION TELEMETRY</p><h2 id="execution-history-title">Recent executions</h2><p className="muted">{cluster?.snapshot.HistoryPersistent ? "Last 128 assigned outcomes, retained across Brain restarts." : "Last 128 assigned outcomes, in memory since Brain startup."} Counters cover this Brain process only. Runtime success is not a durable commit.</p></div></div>
        {cluster?.snapshot.HistoryDegraded && <p role="alert">Execution history storage failed; some completed outcomes may be missing after restart.</p>}
        <div className="neuron-metrics metric-grid">
          {(["Succeeded", "Failed", "Cancelled", "Lost"] as const).map(key => <article key={key}><span>{key}</span><strong>{cluster?.configured ? cluster.snapshot.Executions?.[key] ?? "—" : "—"}</strong></article>)}
        </div>
        <div className="table-wrap neuron-table"><table><thead><tr><th>Execution / trace</th><th>Neuron / runtime</th><th>Outcome</th><th>Wait / elapsed</th><th>CPU / peak RSS</th><th>Finished</th></tr></thead><tbody>
          {cluster?.configured && (cluster.snapshot.Recent ?? []).map((row, index) => <tr key={row.ExecutionID + ":" + row.FinishedAt + ":" + index}><th scope="row">{row.ExecutionID}<small className="table-subtitle muted">{row.TraceID}{row.ParentID ? " · parent " + row.ParentID : ""}</small></th><td>{row.NeuronID}<small className="table-subtitle muted">{row.Runtime.Kind} · {row.Runtime.Profile}</small></td><td>{row.Outcome}</td><td>{row.WaitMillis} / {row.DurationMillis} ms</td><td>{row.CPUTimeNanos > 0 ? (row.CPUTimeNanos / 1e6).toFixed(1) + " ms" : "—"} / {row.PeakMemoryBytes > 0 ? (row.PeakMemoryBytes / 1048576).toFixed(1) + " MiB" : "—"}</td><td><time dateTime={row.FinishedAt}>{new Date(row.FinishedAt).toLocaleString()}</time></td></tr>)}
          {(!cluster?.configured || !cluster.snapshot.Recent?.length) && <tr><td colSpan={6} className="neuron-empty">{cluster?.configured && cluster.snapshot.Executions ? "No completed assignments yet." : "Execution history unavailable."}</td></tr>}
        </tbody></table></div>
      </section>
      <section className="section" aria-labelledby="legacy-workers-title">
        <div className="section-heading"><div><p className="eyebrow">COMPATIBILITY</p><h2 id="legacy-workers-title">Legacy plugin workers</h2><p className="muted">Dedicated workers are separate from shared Neurons.{legacy ? " Project: " + legacy.project : ""}</p></div></div>
        {legacyError && <p role="alert">{legacyError}</p>}
        {!legacy && !loading && !legacyError && <p>Select an active project to inspect dedicated plugin workers. <a href={systemUIURL("projects")}>Choose project</a></p>}
        {legacy && <div className="table-wrap neuron-table"><table><thead><tr><th>Plugin</th><th>Version</th><th>Lifecycle</th><th>Recovery</th></tr></thead><tbody>{legacy.workers.map(worker => <tr key={worker.PluginID + "@" + worker.Version}><th scope="row">{worker.PluginID}</th><td>{worker.Version}</td><td><span className={"status-pill " + (legacyWorkerState(worker) === "Worker ready" ? "is-success" : "is-inactive")}>{legacyWorkerState(worker)}</span></td><td><div className="table-actions"><PluginManageLink projectID={legacy.project} pluginID={worker.PluginID} version={worker.Version} />{worker.Enabled && worker.Ready && !worker.Draining && !worker.WorkerReady && <button type="button" className="secondary" disabled={busy} onClick={() => { setPending(worker); setDone(false); setMessage(""); }}>Start worker</button>}</div></td></tr>)}{!legacy.workers.length && <tr><td colSpan={4} className="neuron-empty">No installed plugin workers.</td></tr>}</tbody></table></div>}
      </section>
      {neuronAction && <AdminDialog title={neuronAction.action + " Neuron"} scope={neuronAction.id} busy={busy} onClose={() => setNeuronAction(undefined)}>
        <p>Apply this cluster-wide action to {neuronAction.id}? Stop and restart wait for active assignments before replacing a process.</p>
        {!done && <button type="button" disabled={busy} onClick={() => void applyNeuronAction()}>Confirm {neuronAction.action}</button>}
        {busy && <AdminProgress label="Applying Neuron action…" />}{done && <p role="status">Neuron action completed</p>}{message && <p role="alert">{message}</p>}
      </AdminDialog>}
      {pending && <AdminDialog title="Start legacy worker" scope={pending.ProjectID + " · " + pending.PluginID + "@" + pending.Version} busy={busy} onClose={() => setPending(undefined)}>{!done && <button type="button" disabled={busy || pending.ProjectID !== legacy?.project} onClick={() => void spinUp()}>Confirm start</button>}{busy && <AdminProgress label="Starting worker…" />}{done && <p role="status">Worker started</p>}{message && <p role="alert">{message}</p>}</AdminDialog>}
    </div>
  </AdminPageFrame>;
}
