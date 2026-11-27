"use client";
import { adminHeaders, readAdminSession, preferredAdminProject, clearAdminSession, signOutAdminSession, projectAdminURL, systemAdminURL, systemUIURL, adminUIContext } from "../lib/admin-api";

import { useEffect, useState } from "react";
import { usePathname, useRouter } from "next/navigation";
import { AdminPageFrame } from "../components/admin-chrome";

type Installation = { PluginID: string; Version: string; Enabled: boolean; Ready: boolean; WorkerReady?: boolean };

export default function WorkersPage() {
  const router = useRouter();
  const { namespace } = adminUIContext(usePathname());
  const [session, setSession] = useState("");
  const [project, setProject] = useState("");
  const [workers, setWorkers] = useState<Installation[]>([]);
  const [message, setMessage] = useState("");

  async function load(token: string) {
    const projectsResponse = await fetch(systemAdminURL("projects"), { headers: adminHeaders(token) });
    if (!projectsResponse.ok) throw new Error("session expired");
    const projects = await projectsResponse.json() as { ID: string; Slug?: string; Active: boolean }[];
    const activeProject = preferredAdminProject(projects, namespace)?.ID ?? "";
    setProject(activeProject);
    if (!activeProject) { setWorkers([]); return; }
    const pluginsResponse = await fetch(projectAdminURL(activeProject, "plugins"), { headers: adminHeaders(token) });
    if (!pluginsResponse.ok) throw new Error("worker status unavailable");
    setWorkers((await pluginsResponse.json() as { installations: Installation[] }).installations);
  }

  useEffect(() => {
    const token = readAdminSession();
    if (!token) { router.replace(systemUIURL()); return; }
    setSession(token);
    void load(token).catch(() => { clearAdminSession(); router.replace(systemUIURL()); });
  }, [router, namespace]);

  async function spinUp(worker: Installation) {
    if (!session || !project) return;
    const response = await fetch(projectAdminURL(project, "actions"), { method: "POST", headers: adminHeaders(session, { "content-type": "application/json" }), body: JSON.stringify({ operation: "plugin.worker.start", project_id: project, plugin_id: worker.PluginID, version: worker.Version }) });
    setMessage(response.ok ? `${worker.PluginID}@${worker.Version} worker started.` : await response.text());
    if (response.ok) await load(session);
  }

  return <AdminPageFrame activePage="workers" title="Workers" projectID={project} onSignOut={() => { void signOutAdminSession().finally(() => router.replace(systemUIURL())); }}><p className="muted">Project: {project || "—"}. A plugin is executable only after its worker is ready.</p>{message && <p className="notice">{message}</p>}<div className="audit-list">{workers.map((worker) => <div className="audit" key={`${worker.PluginID}@${worker.Version}`}><span className={worker.WorkerReady ? "success" : "failure"}>{worker.WorkerReady ? "✓" : "!"}</span><div><strong>{worker.PluginID}@{worker.Version}</strong><small>{worker.WorkerReady ? "Worker ready" : "Worker not ready"}</small></div>{worker.Enabled && worker.Ready && !worker.WorkerReady && <button type="button" onClick={() => void spinUp(worker)}>Spin up worker</button>}</div>)}</div></AdminPageFrame>;
}
