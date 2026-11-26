"use client";
import { adminHeaders, readAdminSession, clearAdminSession, signOutAdminSession } from "../../lib/admin-api";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import { AdminPageFrame } from "../../components/admin-chrome";
import { ModelPanel, type DeclaredModel, type DeclaredModelRoute } from "../../components/model-panel";

type Project = { ID: string; Active: boolean };
type Manifest = { ID: string; Title: string; Description: string; Version: string; Capabilities: string[]; ConfigSchema?: Record<string, string>; Models?: Record<string, DeclaredModel>; Routes?: DeclaredModelRoute[] };
type Installation = { PluginID: string; Version: string; Enabled: boolean; Ready: boolean; Grants: string[] };

export default function PluginSettingsPage() {
  const params = useParams<{ plugin: string }>();
  const router = useRouter();
  const pluginKey = decodeURIComponent(params.plugin ?? "");
  const [session, setSession] = useState("");
  const [project, setProject] = useState<Project | null>(null);
  const [projects, setProjects] = useState<Project[]>([]);
  const [config, setConfig] = useState<Record<string, string>>({});
  const [manifest, setManifest] = useState<Manifest | null>(null);
  const [installation, setInstallation] = useState<Installation | null>(null);
  const [message, setMessage] = useState("");

  async function load(token: string, selectedProject?: string) {
    const projectsResponse = await fetch("/api/admin/projects", { headers: adminHeaders(token) });
    if (!projectsResponse.ok) throw new Error("session expired");
    const projects = await projectsResponse.json() as Project[];
    setProjects(projects.filter((candidate) => candidate.Active));
    const activeProject = projects.find((candidate) => candidate.Active && candidate.ID === selectedProject) ?? projects.find((candidate) => candidate.Active) ?? null;
    setProject(activeProject);
    if (!activeProject) { setManifest(null); setInstallation(null); return; }
    const response = await fetch(`/api/admin/plugins?project_id=${encodeURIComponent(activeProject.ID)}`, { headers: adminHeaders(token) });
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
    const response = await fetch("/api/admin/actions", { method: "POST", headers: adminHeaders(session, { "content-type": "application/json" }), body: JSON.stringify({ operation, project_id: project.ID, plugin_id: installation.PluginID, version: installation.Version, capability, ready: operation === "plugin.readiness", confirmed: operation === "plugin.revoke" }) });
    setMessage(response.ok ? `${operation} completed` : await response.text());
    if (response.ok) await load(session, project.ID);
  }

  async function saveConfig() {
    if (!session || !project || !installation || !manifest) return;
    const values: Record<string, string | number | boolean> = {};
    for (const [field, raw] of Object.entries(config)) {
      const kind = manifest.ConfigSchema?.[field];
      if (kind === "int") {
        const value = Number(raw);
        if (!raw.trim() || !Number.isSafeInteger(value)) { setMessage(field + " must be an integer"); return; }
        values[field] = value;
      } else if (kind === "bool") values[field] = raw === "true";
      else if (kind === "string") values[field] = raw;
    }
    const response = await fetch("/api/admin/actions", { method: "POST", headers: adminHeaders(session, { "content-type": "application/json" }), body: JSON.stringify({ operation: "plugin.configure", project_id: project.ID, plugin_id: installation.PluginID, version: installation.Version, values }) });
    setMessage(response.ok ? "Configuration saved" : await response.text());
    if (response.ok) setConfig({});
  }

  function signOut() {
    void signOutAdminSession().finally(() => router.replace("/"));
  }

  return <AdminPageFrame activePage="plugins" title={manifest?.Title ?? pluginKey} eyebrow="Plugin settings" projectID={project?.ID} onSignOut={signOut}><div className="section-heading"><Link href="/plugins">Back to plugins</Link></div><label>Project <select value={project?.ID ?? ""} onChange={(event) => { setConfig({}); setMessage(""); setInstallation(null); void load(session, event.target.value).catch((error: Error) => setMessage(error.message)); }}>{projects.map((candidate) => <option key={candidate.ID} value={candidate.ID}>{candidate.ID}</option>)}</select></label>{manifest && <><p>{manifest.Description}</p><p><code>{manifest.ID}@{manifest.Version}</code> · {installation?.Ready ? "Ready" : "Not ready"} · {installation?.Enabled ? "Enabled" : "Disabled"}</p><div className="lifecycle-actions">{installation && !installation.Ready && <button onClick={() => lifecycle("plugin.readiness")}>Mark ready</button>}{installation?.Ready && !installation.Enabled && <button onClick={() => lifecycle("plugin.enable")}>Enable</button>}</div>{message && <p className="notice">{message}</p>}{installation && Object.keys(manifest.ConfigSchema ?? {}).length > 0 && <section><h2>Configuration</h2><p className="muted">Only edited fields are saved for the selected project. Existing values are not exposed.</p>{Object.entries(manifest.ConfigSchema ?? {}).map(([name, kind]) => <label key={name}>{name} {kind === "bool" ? <select value={config[name] ?? ""} onChange={(event) => setConfig((old) => { const next = { ...old }; if (event.target.value === "") delete next[name]; else next[name] = event.target.value; return next; })}><option value="">Unchanged</option><option value="true">True</option><option value="false">False</option></select> : <input type={kind === "int" ? "number" : "text"} value={config[name] ?? ""} placeholder="Unchanged" onChange={(event) => setConfig((old) => ({ ...old, [name]: event.target.value }))} />}</label>)}<button disabled={!Object.keys(config).length} onClick={() => void saveConfig()}>Save configuration</button></section>}<h2>Capabilities</h2><div className="capabilities">{manifest.Capabilities.map((capability) => <span key={capability}><code>{capability}</code>{installation?.Grants?.includes(capability) ? <><small>Granted</small><button className="danger" onClick={() => lifecycle("plugin.revoke", capability)}>Revoke</button></> : <button onClick={() => lifecycle("plugin.grant", capability)}>Grant</button>}</span>)}</div>{project && installation?.Enabled && installation.Ready && Object.entries(manifest.Models ?? {}).map(([name, model]) => <ModelPanel key={project.ID + ":" + manifest.ID + ":" + manifest.Version + ":" + name} name={name} model={model} routes={(manifest.Routes ?? []).filter((route) => installation.Grants.includes(route.Capability))} session={session} projectId={project.ID} pluginId={manifest.ID} version={manifest.Version} />)}</>}</AdminPageFrame>;
}
