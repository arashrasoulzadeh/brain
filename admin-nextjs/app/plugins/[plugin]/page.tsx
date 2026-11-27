"use client";
import { adminHeaders, readAdminSession, preferredAdminProject, clearAdminSession, signOutAdminSession, projectAdminURL, systemAdminURL } from "../../lib/admin-api";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import { AdminPageFrame } from "../../components/admin-chrome";
import { AdminConfigForm } from "../../components/admin-config-form";
import { adminContributions } from "../../lib/admin-registry";
import type { SystemManifest, PluginStatus } from "../../lib/admin-types";

type Project = { ID: string; Active: boolean };
type Manifest = SystemManifest;
type Installation = PluginStatus;

export default function PluginSettingsPage() {
  const params = useParams<{ plugin: string }>();
  const router = useRouter();
  const pluginKey = decodeURIComponent(params.plugin ?? "");
  const [session, setSession] = useState("");
  const [project, setProject] = useState<Project | null>(null);
  const [manifest, setManifest] = useState<Manifest | null>(null);
  const [installation, setInstallation] = useState<Installation | null>(null);
  const [message, setMessage] = useState("");

  async function load(token: string) {
    const projectsResponse = await fetch(systemAdminURL("projects"), { headers: adminHeaders(token) });
    if (!projectsResponse.ok) throw new Error("session expired");
    const projects = await projectsResponse.json() as Project[];
    const activeProject = preferredAdminProject(projects) ?? null;
    setProject(activeProject);
    if (!activeProject) { setManifest(null); setInstallation(null); return; }
    const response = await fetch(projectAdminURL(activeProject.ID, "plugins"), { headers: adminHeaders(token) });
    if (!response.ok) throw new Error("plugin data unavailable");
    const payload = await response.json() as { manifests: Manifest[]; installations: Installation[] };
    const [pluginID, version] = pluginKey.split("@");
    setManifest(payload.manifests.find((candidate) => candidate.ID === pluginID && candidate.Version === version) ?? null);
    setInstallation(payload.installations.find((candidate) => candidate.PluginID === pluginID && candidate.Version === version) ?? null);
  }

  useEffect(() => {
    const token = readAdminSession();
    if (!token) { router.replace("/"); return; }
    setSession(token);
    void load(token).catch(() => { clearAdminSession(); router.replace("/"); });
  }, [pluginKey, router]);

  async function lifecycle(operation: string, capability = "") {
    if (!session || !project || !installation) return;
    if (operation === "plugin.revoke" && !window.confirm(`Revoke ${capability}?`)) return;
    const response = await fetch(operation === "plugin.readiness" ? systemAdminURL("actions") : projectAdminURL(project.ID, "actions"), { method: "POST", headers: adminHeaders(session, { "content-type": "application/json" }), body: JSON.stringify({ operation, project_id: project.ID, plugin_id: installation.PluginID, version: installation.Version, capability, ready: operation === "plugin.readiness", confirmed: operation === "plugin.revoke" }) });
    setMessage(response.ok ? `${operation} completed` : await response.text());
    if (response.ok) await load(session);
  }

  function signOut() {
    void signOutAdminSession().finally(() => router.replace("/"));
  }

  return <AdminPageFrame activePage="plugins" title={manifest?.Title ?? pluginKey} eyebrow="Plugin settings" projectID={project?.ID} onSignOut={signOut}><div className="section-heading"><Link href="/plugins">Back to plugins</Link></div><p className="muted">Project: {project?.ID ?? "—"}</p>{manifest && <><p>{manifest.Description}</p><p><code>{manifest.ID}@{manifest.Version}</code> · {installation?.Ready ? "Ready" : "Not ready"} · {installation?.Enabled ? "Enabled" : "Disabled"}</p><div className="lifecycle-actions">{installation && !installation.Ready && <button onClick={() => lifecycle("plugin.readiness")}>Mark ready</button>}{installation?.Ready && !installation.Enabled && <button onClick={() => lifecycle("plugin.enable")}>Enable</button>}</div>{message && <p className="notice">{message}</p>}<h2>Capabilities</h2><div className="capabilities">{manifest.Capabilities.map((capability) => <span key={capability}><code>{capability}</code>{installation?.Grants?.includes(capability) ? <><small>Granted</small><button className="danger" onClick={() => lifecycle("plugin.revoke", capability)}>Revoke</button></> : <button onClick={() => lifecycle("plugin.grant", capability)}>Grant</button>}</span>)}</div>{project && installation && manifest.ConfigSchema && <AdminConfigForm key={project.ID + ":" + manifest.ID + "@" + manifest.Version} token={session} projectID={project.ID} pluginID={manifest.ID} version={manifest.Version} schema={manifest.ConfigSchema} />}{project && installation && <section className="section"><h2>Admin pages</h2>{adminContributions(project.ID, { manifests: [manifest], installations: [installation] }).map(item => <p key={item.href}><Link href={item.href}>{item.page.title}</Link></p>)}</section>}</>}</AdminPageFrame>;
}
