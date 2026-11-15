"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { AdminFooter, AdminHeader, AdminSidebar } from "../components/admin-chrome";

type Installation = { PluginID: string; Version: string; Enabled: boolean; Ready: boolean; WorkerReady?: boolean };

export default function WorkersPage() {
  const router = useRouter();
  const [session, setSession] = useState("");
  const [project, setProject] = useState("");
  const [workers, setWorkers] = useState<Installation[]>([]);
  const [message, setMessage] = useState("");

  async function load(token: string) {
    const projectsResponse = await fetch("/api/admin/projects", { headers: { authorization: `Bearer ${token}` } });
    if (!projectsResponse.ok) throw new Error("session expired");
    const projects = await projectsResponse.json() as { ID: string; Active: boolean }[];
    const activeProject = projects.find((candidate) => candidate.Active)?.ID ?? "";
    setProject(activeProject);
    if (!activeProject) return;
    const pluginsResponse = await fetch(`/api/admin/plugins?project_id=${encodeURIComponent(activeProject)}`, { headers: { authorization: `Bearer ${token}` } });
    if (!pluginsResponse.ok) throw new Error("worker status unavailable");
    setWorkers((await pluginsResponse.json() as { installations: Installation[] }).installations);
  }

  useEffect(() => {
    const token = window.sessionStorage.getItem("brain-admin-session") ?? "";
    if (!token) { router.replace("/"); return; }
    setSession(token);
    void load(token).catch(() => { window.sessionStorage.removeItem("brain-admin-session"); router.replace("/"); });
  }, [router]);

  async function spinUp(worker: Installation) {
    const response = await fetch("/api/admin/actions", { method: "POST", headers: { authorization: `Bearer ${session}`, "content-type": "application/json" }, body: JSON.stringify({ operation: "plugin.worker.start", project_id: project, plugin_id: worker.PluginID, version: worker.Version }) });
    setMessage(response.ok ? `${worker.PluginID}@${worker.Version} worker started.` : await response.text());
    if (response.ok) await load(session);
  }

  return <main className="shell"><div className="workspace"><AdminSidebar activePage="workers" projectID={project} onSignOut={() => { window.sessionStorage.removeItem("brain-admin-session"); router.replace("/"); }} /><section className="card admin-content"><AdminHeader title="Workers" /><p className="muted">Project: {project || "—"}. A plugin is executable only after its worker is ready.</p>{message && <p className="notice">{message}</p>}<div className="audit-list">{workers.map((worker) => <div className="audit" key={`${worker.PluginID}@${worker.Version}`}><span className={worker.WorkerReady ? "success" : "failure"}>{worker.WorkerReady ? "✓" : "!"}</span><div><strong>{worker.PluginID}@{worker.Version}</strong><small>{worker.WorkerReady ? "Worker ready" : "Worker not ready"}</small></div>{worker.Enabled && worker.Ready && !worker.WorkerReady && <button type="button" onClick={() => void spinUp(worker)}>Spin up worker</button>}</div>)}</div><AdminFooter /></section></div></main>;
}
