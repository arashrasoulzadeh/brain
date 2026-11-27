"use client";
import { useEffect, useState } from "react";
import { usePathname } from "next/navigation";
import Link from "next/link";
import { AdminPageFrame } from "../components/admin-chrome";
import { AdminProgress } from "../components/admin-dialog";
import { PluginManageLink } from "../components/plugin-manage-link";
import { adminHeaders, readAdminSession, preferredAdminProject, projectAdminURL, projectUIURL, systemAdminURL, systemUIURL, adminUIContext, ADMIN_PLUGIN_CATALOG_CHANGED } from "../lib/admin-api";
import type { AdminCatalog } from "../lib/admin-registry";
import type { Project } from "../lib/admin-types";

export default function DashboardWorkspace() {
  const { namespace } = adminUIContext(usePathname());
  const [epoch, setEpoch] = useState(0);
  const [data, setData] = useState<{ project: Project; catalog: AdminCatalog }>();
  const [error, setError] = useState("");
  useEffect(() => {
    const changed = () => setEpoch(value => value + 1);
    window.addEventListener(ADMIN_PLUGIN_CATALOG_CHANGED, changed);
    return () => window.removeEventListener(ADMIN_PLUGIN_CATALOG_CHANGED, changed);
  }, []);
  useEffect(() => {
    const controller = new AbortController();
    setData(undefined); setError("");
    void (async () => {
      const token = readAdminSession();
      if (!token || !namespace) throw new Error("Select an active project");
      const options = { headers: adminHeaders(token), signal: controller.signal };
      const response = await fetch(systemAdminURL("projects"), options);
      if (!response.ok) throw new Error("Projects unavailable");
      const project = preferredAdminProject(await response.json() as Project[], namespace);
      if (!project) throw new Error("Project unavailable");
      const plugins = await fetch(projectAdminURL(project.ID, "plugins"), options);
      if (!plugins.ok) throw new Error("Plugin catalog unavailable");
      const catalog = await plugins.json() as AdminCatalog;
      if (!controller.signal.aborted) setData({ project, catalog });
    })().catch(reason => { if (!controller.signal.aborted) setError(String(reason)); });
    return () => controller.abort();
  }, [namespace, epoch]);
  const installations = data ? (data.catalog.installations ?? []).filter(item => item.ProjectID === data.project.ID) : [];
  const dashboard = installations.filter(item => item.PluginID === "dashboard" && item.Enabled && item.Ready && !item.Draining);
  return <AdminPageFrame activePage="dashboard" title="Dashboard" requiredPluginID="dashboard">
    {error ? <p role="alert">{error}</p> : !data ? <AdminProgress label="Loading project dashboard…" /> : <>
      <section className="section"><div className="section-heading"><h2>{data.project.Name || data.project.ID}</h2><div className="lifecycle-actions">{dashboard.map(item => <PluginManageLink key={item.Version} projectID={data.project.ID} pluginID={item.PluginID} version={item.Version} />)}</div></div>
        <p>{installations.length} installed · {installations.filter(item => item.Enabled && item.Ready && !item.Draining).length} enabled · {installations.filter(item => item.WorkerReady).length} running</p>
        <div className="lifecycle-actions"><Link href={projectUIURL(data.project.ID, "plugins")}>Plugin catalog</Link><Link href={projectUIURL(data.project.ID, "workers")}>Workers</Link><Link href={projectUIURL(data.project.ID, "routes")}>API routes</Link><Link href={systemUIURL("audit")}>Audit log</Link></div>
      </section>
      <section className="section"><h2>Project installations</h2><div className="table-wrap"><table><thead><tr><th>Plugin</th><th>Version</th><th>Status</th><th>Settings</th></tr></thead><tbody>{installations.map(item => <tr key={item.PluginID + "@" + item.Version}><td>{item.PluginID}</td><td>{item.Version}</td><td>{item.Draining ? "Draining" : item.Enabled && item.Ready ? "Enabled" : "Disabled"}</td><td><PluginManageLink projectID={data.project.ID} pluginID={item.PluginID} version={item.Version} /></td></tr>)}</tbody></table></div></section>
    </>}
  </AdminPageFrame>;
}
