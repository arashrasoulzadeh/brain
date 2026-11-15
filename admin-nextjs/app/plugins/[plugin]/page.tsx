"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";

type Project = { ID: string; Active: boolean };
type Manifest = { ID: string; Title: string; Description: string; Version: string; Capabilities: string[]; ConfigSchema?: Record<string, string> };
type Installation = { ProjectID: string; PluginID: string; Version: string; Enabled: boolean; Ready: boolean; Draining: boolean };

export default function PluginSettingsPage() {
  const params = useParams<{ plugin: string }>();
  const pluginKey = decodeURIComponent(params.plugin ?? "");
  const [session, setSession] = useState("");
  const [project, setProject] = useState<Project | null>(null);
  const [manifest, setManifest] = useState<Manifest | null>(null);
  const [installation, setInstallation] = useState<Installation | null>(null);
  const [message, setMessage] = useState("");

  async function load(token: string) {
    const projectsResponse = await fetch("/api/admin/projects", { headers: { authorization: `Bearer ${token}` } });
    const projects = await projectsResponse.json() as Project[];
    const selected = projects.find((candidate) => candidate.Active) ?? null;
    setProject(selected);
    if (!selected) return;
    const response = await fetch(`/api/admin/plugins?project_id=${encodeURIComponent(selected.ID)}`, { headers: { authorization: `Bearer ${token}` } });
    const payload = await response.json() as { manifests: Manifest[]; installations: Installation[] };
    const [pluginID, version] = pluginKey.split("@");
    setManifest(payload.manifests.find((candidate) => candidate.ID === pluginID && candidate.Version === version) ?? null);
    setInstallation(payload.installations.find((candidate) => candidate.PluginID === pluginID && candidate.Version === version) ?? null);
  }

  useEffect(() => {
    const token = window.sessionStorage.getItem("brain-admin-session") ?? "";
    setSession(token);
    if (token) void load(token);
  }, [pluginKey]);

  async function lifecycle(operation: string) {
    if (!session || !project || !installation) return;
    const response = await fetch("/api/admin/actions", { method: "POST", headers: { authorization: `Bearer ${session}`, "content-type": "application/json" }, body: JSON.stringify({ operation, project_id: project.ID, plugin_id: installation.PluginID, version: installation.Version }) });
    setMessage(response.ok ? `${operation} completed` : await response.text());
    if (response.ok) await load(session);
  }

  if (!session) return <main className="shell"><section className="section"><h1>Sign in required</h1><a href="/plugins">Return to plugins</a></section></main>;
  return <main className="shell"><section className="section"><div className="section-heading"><div><p className="eyebrow">Plugin settings</p><h1>{manifest?.Title ?? pluginKey}</h1><p className="muted">Project: {project?.ID ?? "—"}</p></div><a href="/plugins">Back to plugins</a></div>{manifest && <><p>{manifest.Description}</p><p><code>{manifest.ID}@{manifest.Version}</code> · {installation?.Ready ? "Ready" : "Not ready"} · {installation?.Enabled ? "Enabled" : "Disabled"}</p><div className="lifecycle-actions">{installation && !installation.Ready && <button onClick={() => lifecycle("plugin.readiness")}>Mark ready</button>}{installation?.Ready && !installation.Enabled && <button onClick={() => lifecycle("plugin.enable")}>Enable</button>}</div>{message && <p className="muted">{message}</p>}<h2>Capabilities</h2><div className="capabilities">{manifest.Capabilities.map((capability) => <span key={capability}>{capability}</span>)}</div></>}</section></main>;
}
