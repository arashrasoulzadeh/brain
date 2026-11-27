"use client";
import { adminHeaders, readAdminSession, preferredAdminProject, signOutAdminSession, projectAdminURL, systemAdminURL, projectUIURL, systemUIURL, adminUIContext, notifyPluginCatalogChanged } from "../../lib/admin-api";
import { useEffect, useRef, useState } from "react";
import { useParams, usePathname, useRouter } from "next/navigation";
import Link from "next/link";
import { AdminPageFrame } from "../../components/admin-chrome";
import { AdminConfigForm } from "../../components/admin-config-form";
import { AdminDialog, AdminProgress } from "../../components/admin-dialog";
import { PluginInstallDialog } from "../../components/plugin-install-dialog";
import { installAndActivatePlugin } from "../../lib/plugin-install";
import { adminContributions } from "../../lib/admin-registry";
import type { SystemManifest, PluginStatus } from "../../lib/admin-types";

type Project = { ID: string; Slug?: string; Active: boolean };
type PendingAction = { operation: string; label: string; capability?: string; resource?: string; preview?: { Revision: number; Dependents: string[] } };

export default function PluginSettingsPage() {
  const params = useParams<{ plugin: string }>();
  const { namespace } = adminUIContext(usePathname());
  const pluginKey = decodeURIComponent(params.plugin ?? "");
  return <ProjectPluginSettings key={(namespace ?? "") + ":" + pluginKey} namespace={namespace} pluginKey={pluginKey} />;
}

function ProjectPluginSettings({ namespace, pluginKey }: { namespace?: string; pluginKey: string }) {
  const router = useRouter();
  const [session, setSession] = useState("");
  const [project, setProject] = useState<Project | null>(null);
  const [manifest, setManifest] = useState<SystemManifest | null>(null);
  const [installation, setInstallation] = useState<PluginStatus | null>(null);
  const [message, setMessage] = useState("");
  const [loading, setLoading] = useState(true);
  const [installOpen, setInstallOpen] = useState(false);
  const [pendingAction, setPendingAction] = useState<PendingAction>();
  const [busy, setBusy] = useState(false);
  const [actionError, setActionError] = useState("");
  const [actionDone, setActionDone] = useState(false);
  const loadEpoch = useRef(0);

  async function load(token: string) {
    const epoch = ++loadEpoch.current;
    const current = () => epoch === loadEpoch.current && readAdminSession() === token;
    const headers = adminHeaders(token);
    const projectsResponse = await fetch(systemAdminURL("projects"), { headers });
    if (!projectsResponse.ok) throw new Error("Projects unavailable");
    const projects = await projectsResponse.json() as Project[];
    if (!current()) return;
    const activeProject = namespace ? preferredAdminProject(projects, namespace) : undefined;
    if (!activeProject) { setProject(null); setManifest(null); setInstallation(null); throw new Error("Project unavailable"); }
    const response = await fetch(projectAdminURL(activeProject.ID, "plugins"), { headers });
    if (!response.ok) throw new Error("Plugin data unavailable");
    const payload = await response.json() as { manifests: SystemManifest[] | null; installations: PluginStatus[] | null };
    if (!current()) return;
    const [pluginID, version] = pluginKey.split("@");
    const registered = (payload.manifests ?? []).find(candidate => candidate.ID === pluginID && candidate.Version === version);
    setProject(activeProject); setManifest(registered ?? null);
    setInstallation((payload.installations ?? []).find(candidate => candidate.ProjectID === activeProject.ID && candidate.PluginID === pluginID && candidate.Version === version) ?? null);
    if (!registered) setMessage("Plugin version unavailable");
  }

  useEffect(() => {
    const token = readAdminSession();
    if (!token) { router.replace(systemUIURL()); return; }
    setSession(token); setLoading(true); setMessage("");
    let active = true;
    void load(token).catch(reason => { if (active && readAdminSession() === token) setMessage(String(reason)); }).finally(() => { if (active) setLoading(false); });
    return () => { active = false; loadEpoch.current++; };
  }, [router, namespace, pluginKey]);

  async function command(operation: string, extra: Record<string, unknown> = {}) {
    if (!project || !manifest || !session) throw new Error("Select an active project and plugin version");
    const response = await fetch(operation === "plugin.readiness" ? systemAdminURL("actions") : projectAdminURL(project.ID, "actions"), {
      method: "POST", headers: adminHeaders(session, { "content-type": "application/json" }),
      body: JSON.stringify({ operation, project_id: project.ID, plugin_id: manifest.ID, version: manifest.Version, ...extra })
    });
    if (!response.ok) throw new Error(await response.text());
    return response;
  }

  async function openAction(operation: string, label: string, capability?: string) {
    if (busy) return;
    setPendingAction({ operation, label, capability }); setActionDone(false); setActionError("");
    if (operation === "plugin.disable.apply") {
      setBusy(true);
      try {
        const response = await command("plugin.disable.preview");
        const preview = await response.json() as { Revision: number; Dependents: string[] };
        if (!Number.isSafeInteger(preview.Revision) || !Array.isArray(preview.Dependents)) throw new Error("Invalid dependency preview");
        setPendingAction({ operation, label, preview });
      } catch (reason) { setActionError(String(reason)); }
      finally { setBusy(false); }
    }
  }

  async function lifecycle() {
    if (busy || actionDone || !pendingAction || !project || !manifest) return;
    const { operation, capability, resource, preview } = pendingAction;
    if (operation === "plugin.grant-resource" && !resource?.trim()) return;
    if (operation !== "plugin.install" && !installation) return;
    if (operation === "plugin.disable.apply" && !preview) return;
    setBusy(true); setActionError("");
    try {
      await command(operation, { capability, resource, ready: operation === "plugin.readiness", confirmed: true, ...(preview ? { revision: preview.Revision, dependents: preview.Dependents } : {}) });
      setActionDone(true);
      notifyPluginCatalogChanged(project.ID);
      await load(session).catch(reason => setActionError("Action completed; refresh failed: " + String(reason)));
    } catch (reason) { setActionError(String(reason)); notifyPluginCatalogChanged(project.ID); }
    finally { setBusy(false); }
  }

  const scope = (project?.ID ?? "") + " · " + pluginKey;
  function signOut() { void signOutAdminSession().finally(() => router.replace(systemUIURL())); }
  const actionButton = (operation: string, label: string, danger = false) => <button type="button" className={danger ? "danger" : "secondary"} disabled={busy} onClick={() => void openAction(operation, label)}>{label}</button>;
  return <AdminPageFrame activePage="plugins" title={manifest?.Title ?? pluginKey} eyebrow="Project plugin management" projectID={project?.ID} onSignOut={signOut}>
    <div className="section-heading"><Link href={project ? projectUIURL(project.ID, "plugins") : systemUIURL("projects")}>Back to plugins</Link></div>
    <p className="muted">Project: {project?.ID ?? "—"}</p>
    {loading && <AdminProgress label="Loading project installation…" />}
    {message && <p role="alert">{message}</p>}
    {installOpen && project && manifest && <PluginInstallDialog projectID={project.ID} manifest={manifest} onClose={() => setInstallOpen(false)} onInstall={async () => { await installAndActivatePlugin(session, project.ID, manifest); await load(session).catch(reason => setMessage("Plugin installed; refresh failed: " + String(reason))); }} />}
    {pendingAction && project && manifest && <AdminDialog title={pendingAction.label} scope={scope} busy={busy} onClose={() => setPendingAction(undefined)}>
      <p>{pendingAction.capability || (pendingAction.operation === "plugin.readiness" ? "Artifact readiness is shared across projects." : "This action applies only to the selected installation.")}</p>
      {pendingAction.operation === "plugin.grant-resource" && <><p className="muted">The first exact resource grant restricts resource-bound operations for this capability to the listed resources. Further grants add exact keys. Other access policies still apply.</p><label>Exact resource key<input disabled={busy || actionDone} value={pendingAction.resource ?? ""} onChange={event => setPendingAction({ ...pendingAction, resource: event.target.value })} /></label></>}
      {pendingAction.operation === "plugin.stop-immediate" && <p role="alert">Stop accepting work immediately; active work may be interrupted.</p>}
      {pendingAction.operation === "plugin.rollback" && <p>Make this exact version the active version. The host checks compatibility, grants and configuration before accepting it.</p>}
      {pendingAction.operation === "plugin.recover" && <p>Clear the draining or stopped state. Start the worker separately if required.</p>}
      {pendingAction.preview && <p>{pendingAction.preview.Dependents.length ? "Affected dependencies: " + pendingAction.preview.Dependents.join(", ") : "No affected dependencies."}</p>}
      {!actionDone && <button type="button" disabled={busy || (pendingAction.operation === "plugin.disable.apply" && !pendingAction.preview) || (pendingAction.operation === "plugin.grant-resource" && !pendingAction.resource?.trim())} onClick={() => void lifecycle()}>Confirm action</button>}
      {busy && <AdminProgress label="Applying change…" />}{actionDone && <p role="status">Action completed</p>}{actionError && <p role="alert">{actionError}</p>}
    </AdminDialog>}
    {manifest && project && <>
      <section className="section admin-action-panel"><h2>Installation</h2><p>{manifest.Description}</p><p><code>{manifest.ID}@{manifest.Version}</code> · {!installation ? "Not installed" : installation.Draining ? "Draining" : installation.Enabled ? "Enabled" : "Disabled"} · {installation?.Ready ? "Artifact ready" : "Artifact not ready"} · {installation?.WorkerReady ? "Worker ready" : "Worker not ready"}</p>
        <div className="lifecycle-actions">
          {!installation ? <button type="button" onClick={() => setInstallOpen(true)}>Install</button> : <>
            {!installation.Ready && actionButton("plugin.readiness", "Mark artifact ready")}
            {installation.Ready && !installation.Enabled && !installation.Draining && actionButton("plugin.enable", "Enable")}
            {installation.Enabled && installation.Ready && !installation.WorkerReady && actionButton("plugin.worker.start", "Start worker")}
            {installation.Enabled && !installation.Draining && actionButton("plugin.drain", "Drain")}
            {installation.Draining && actionButton("plugin.resume", "Resume")}
            {installation.Enabled && actionButton("plugin.stop-immediate", "Stop immediately", true)}
            {installation.Draining && actionButton("plugin.recover", "Recover")}
            {installation.Enabled && installation.Ready && !installation.Draining && actionButton("plugin.rollback", "Roll back to this version", true)}
            {installation.Enabled && actionButton("plugin.disable.apply", "Disable", true)}
            {!installation.Enabled && !installation.Draining && actionButton("plugin.remove", "Uninstall", true)}
            {installation.Enabled && actionButton("plugin.promote", "Promote version")}
          </>}
        </div>
      </section>
      {installation ? <>
        <AdminConfigForm key={scope} token={session} projectID={project.ID} pluginID={manifest.ID} version={manifest.Version} schema={manifest.ConfigSchema ?? {}} configuredFields={installation.ConfiguredFields} publicConfig={installation.PublicConfig} maxUploadBytes={installation.MaxUploadBytes} onSaved={() => load(session)} />
        <section className="section admin-action-panel"><h2>Capabilities</h2><p className="muted">Capability grants do not override explicit project access policies. Resource keys are exact, not wildcard patterns.</p><div className="table-wrap"><table><thead><tr><th>Capability</th><th>Scope</th><th>Actions</th></tr></thead><tbody>{manifest.Capabilities.map(capability => {
          const granted = installation.Grants?.includes(capability);
          const resources = installation.ResourceGrants?.[capability] ?? [];
          const supportsResource = manifest.Routes?.some(route => route.Capability === capability && route.ResourceField);
          return <tr key={capability}><td><code>{capability}</code></td><td>{!granted ? "Not granted" : resources.length ? <ul>{resources.map(resource => <li key={resource}><code>{resource}</code></li>)}</ul> : "Capability-wide"}</td><td><div className="lifecycle-actions">
            {granted ? <button type="button" className="danger" disabled={busy} onClick={() => void openAction("plugin.revoke", "Revoke capability and resource grants", capability)}>Revoke</button> : <button type="button" disabled={busy} onClick={() => void openAction("plugin.grant", "Grant capability", capability)}>Grant</button>}
            {supportsResource && <button type="button" className="secondary" disabled={busy} onClick={() => void openAction("plugin.grant-resource", "Grant exact resource", capability)}>Grant resource</button>}
          </div></td></tr>;
        })}</tbody></table></div>{!manifest.Capabilities.length && <p className="muted">No declared capabilities.</p>}</section>
        <section className="section"><h2>Admin pages</h2>{adminContributions(project.ID, { manifests: [manifest], installations: [installation] }).map(item => <p key={item.href}><Link href={item.href}>{item.page.title}</Link></p>)}{manifest.ID === "dashboard" && <p><Link href={systemUIURL("projects")}>Project control plane</Link> · <Link href={projectUIURL(project.ID, "plugins")}>Plugin catalog</Link></p>}{manifest.ID === "functions" && <p><Link href={projectUIURL(project.ID, "functions")}>Function editor and publications</Link></p>}</section>
      </> : <p className="muted">Install this plugin in {project.ID} to edit settings and capabilities.</p>}
    </>}
  </AdminPageFrame>;
}
