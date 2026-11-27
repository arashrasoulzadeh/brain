"use client";
import { adminHeaders, readAdminSession, preferredAdminProject, signOutAdminSession, projectAdminURL, systemAdminURL, systemUIURL, adminUIContext, notifyPluginCatalogChanged } from "../lib/admin-api";

import { useEffect, useRef, useState } from "react";
import { usePathname, useRouter } from "next/navigation";
import { AdminDialog, AdminProgress } from "../components/admin-dialog";
import { PluginManageLink } from "../components/plugin-manage-link";
import { AdminPageFrame } from "../components/admin-chrome";

type Installation = { ProjectID: string; Draining?: boolean; PluginID: string; Version: string; Enabled: boolean; Ready: boolean; WorkerReady?: boolean };

export default function WorkersPage() {
  const router = useRouter();
  const { namespace } = adminUIContext(usePathname());
  const [session, setSession] = useState("");
  const [project, setProject] = useState("");
  const [workers, setWorkers] = useState<Installation[]>([]);
  const [message, setMessage] = useState("");
  const [loading, setLoading] = useState(true);
  const [pending, setPending] = useState<Installation>();
  const [busy, setBusy] = useState(false);
  const [done, setDone] = useState(false);
  const epoch = useRef(0);

  async function load(token: string, signal?: AbortSignal) {
    const current = ++epoch.current;
    const valid = () => current === epoch.current && !signal?.aborted && readAdminSession() === token;
    const options = { headers: adminHeaders(token), signal };
    const projectsResponse = await fetch(systemAdminURL("projects"), options);
    if (!projectsResponse.ok) throw new Error("Projects unavailable");
    const projects = (await projectsResponse.json() as { ID: string; Slug?: string; Active: boolean }[] | null) ?? [];
    const activeProject = preferredAdminProject(projects, namespace)?.ID;
    if (!activeProject) throw new Error("Project unavailable");
    const pluginsResponse = await fetch(projectAdminURL(activeProject, "plugins"), options);
    if (!pluginsResponse.ok) throw new Error("Worker status unavailable");
    const data = await pluginsResponse.json() as { installations: Installation[] | null };
    if (valid()) { setProject(activeProject); setWorkers((data.installations ?? []).filter(worker => worker.ProjectID === activeProject)); }
  }

  useEffect(() => {
    const controller = new AbortController();
    const token = readAdminSession();
    if (!token) { router.replace(systemUIURL()); return; }
    setSession(token); setProject(""); setWorkers([]); setMessage(""); setLoading(true); setPending(undefined);
    void load(token, controller.signal).catch(reason => { if (!controller.signal.aborted) setMessage(String(reason)); }).finally(() => { if (!controller.signal.aborted) setLoading(false); });
    return () => { controller.abort(); epoch.current++; };
  }, [router, namespace]);

  async function spinUp() {
    if (busy || done || !pending || pending.ProjectID !== project || !session) return;
    setBusy(true); setMessage("");
    try {
      const response = await fetch(projectAdminURL(project, "actions"), { method: "POST", headers: adminHeaders(session, { "content-type": "application/json" }), body: JSON.stringify({ operation: "plugin.worker.start", project_id: project, plugin_id: pending.PluginID, version: pending.Version }) });
      if (!response.ok) throw new Error(await response.text());
      setDone(true); notifyPluginCatalogChanged(project); await load(session);
    } catch (reason) { setMessage(String(reason)); }
    finally { setBusy(false); }
  }

  return <AdminPageFrame activePage="workers" title="Workers" projectID={project} onSignOut={() => { void signOutAdminSession().finally(() => router.replace(systemUIURL())); }}><p className="muted">Project: {project || "—"}. A plugin is executable only after its worker is ready.</p>{loading && <AdminProgress label="Loading workers…" />}{!pending && message && <p role="alert">{message}</p>}{!loading && !workers.length && !message && <p>No installed workers.</p>}<div className="audit-list">{workers.map(worker => <div className="audit" key={worker.PluginID + "@" + worker.Version}><span className={worker.WorkerReady ? "success" : "failure"}>{worker.WorkerReady ? "✓" : "!"}</span><div><strong>{worker.PluginID}@{worker.Version}</strong><small>{worker.Draining ? "Draining" : worker.WorkerReady ? "Worker ready" : "Worker not ready"}</small></div><div className="lifecycle-actions"><PluginManageLink projectID={project} pluginID={worker.PluginID} version={worker.Version} />{worker.Enabled && worker.Ready && !worker.Draining && !worker.WorkerReady && <button type="button" onClick={() => { setPending(worker); setDone(false); setMessage(""); }}>Start worker</button>}</div></div>)}</div>{pending && <AdminDialog title="Start worker" scope={project + " · " + pending.PluginID + "@" + pending.Version} busy={busy} onClose={() => setPending(undefined)}>{!done && <button type="button" disabled={busy} onClick={() => void spinUp()}>Confirm start</button>}{busy && <AdminProgress label="Starting worker…" />}{done && <p role="status">Worker started</p>}{message && <p role="alert">{message}</p>}</AdminDialog>}</AdminPageFrame>;
}
