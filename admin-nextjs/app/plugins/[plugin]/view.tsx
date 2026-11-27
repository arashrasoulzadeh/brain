"use client";
import { adminHeaders, readAdminSession, preferredAdminProject, signOutAdminSession, projectAdminURL, systemAdminURL, projectUIURL, systemUIURL, adminUIContext, notifyPluginCatalogChanged } from "../../lib/admin-api";
import { useEffect, useRef, useState } from "react";
import { useParams, usePathname, useRouter } from "next/navigation";
import Link from "next/link";
import { AdminPageFrame } from "../../components/admin-chrome";
import { AdminConfigForm } from "../../components/admin-config-form";
import { AdminDialog, AdminProgress } from "../../components/admin-dialog";
import { PluginInstallDialog } from "../../components/plugin-install-dialog";
import { PluginEnableDependencies } from "../../components/plugin-enable-dependencies";
import { installAndActivatePlugin } from "../../lib/plugin-install";
import { dependencyEnablePlan, type DependencyEnablePlan, pluginVersionReview, pluginCatalogState } from "../../lib/plugin-version-review";
import { adminContributions } from "../../lib/admin-registry";
import type { SystemManifest, PluginStatus } from "../../lib/admin-types";

type Project = { ID: string; Slug?: string; Active: boolean };
type PendingAction = { operation: string; label: string; capability?: string; resource?: string; dependencyPlan?: DependencyEnablePlan; preview?: { Revision: number; Dependents: string[] } };

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
  const [catalog, setCatalog] = useState<{ manifests: SystemManifest[]; installations: PluginStatus[] }>({ manifests: [], installations: [] });
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
    setCatalog({ manifests: payload.manifests ?? [], installations: (payload.installations ?? []).filter(item => item.ProjectID === activeProject.ID) });
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
    if (!project || !manifest || !session || session !== readAdminSession()) throw new Error("Select an active project and plugin version");
    const response = await fetch(operation === "plugin.readiness" ? systemAdminURL("actions") : projectAdminURL(project.ID, "actions"), {
      method: "POST", headers: adminHeaders(session, { "content-type": "application/json" }),
      body: JSON.stringify({ operation, project_id: project.ID, plugin_id: manifest.ID, version: manifest.Version, ...extra })
    });
    if (!response.ok) throw new Error(await response.text());
    return response;
  }

  async function openAction(operation: string, label: string, capability?: string) {
    if (busy) return;
    setPendingAction({ operation, label, capability, ...(operation === "plugin.enable" && project && manifest ? { dependencyPlan: dependencyEnablePlan(project.ID, manifest, catalog.manifests, catalog.installations) } : {}) }); setActionDone(false); setActionError("");
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
    const { operation, capability, resource, preview, dependencyPlan } = pendingAction;
    if (operation === "plugin.enable" && (!dependencyPlan || dependencyPlan.blockers.length)) return;
    if (operation === "plugin.grant-resource" && !resource?.trim()) return;
    if (operation !== "plugin.install" && !installation) return;
    if (operation === "plugin.disable.apply" && !preview) return;
    setBusy(true); setActionError("");
    try {
      await command(operation, { capability, resource, ready: operation === "plugin.readiness", confirmed: true, ...(dependencyPlan ? { dependencies: dependencyPlan.enable } : {}), ...(preview ? { revision: preview.Revision, dependents: preview.Dependents } : {}) });
      setActionDone(true);
      notifyPluginCatalogChanged(project.ID);
      await load(session).catch(reason => setActionError("Action completed; refresh failed: " + String(reason)));
    } catch (reason) { setActionError(String(reason)); notifyPluginCatalogChanged(project.ID); }
    finally { setBusy(false); }
  }

  const scope = (project?.ID ?? "") + " · " + pluginKey;
  const review = manifest && project ? pluginVersionReview(project.ID, manifest, catalog.manifests, catalog.installations) : undefined;
  const canPromote = !!installation?.Enabled && installation.Ready && !installation.Draining && !!review && !review.missingGrants.length && !review.missingApprovals.length && !review.missingConfig.length && !review.incompatibleConfig.length && !review.requiredUnavailable.length;
  const enablePlan = manifest && project ? dependencyEnablePlan(project.ID, manifest, catalog.manifests, catalog.installations) : undefined;
  const canActivateFresh = !!review && !review.otherEnabled.length && !review.promoted && !review.requiredUnavailable.length;
  function signOut() { void signOutAdminSession().finally(() => router.replace(systemUIURL())); }
  const actionButton = (operation: string, label: string, danger = false) => <button type="button" className={danger ? "danger" : "secondary"} disabled={busy} onClick={() => void openAction(operation, label)}>{label}</button>;
  return <AdminPageFrame activePage="plugins" title={manifest?.Title ?? pluginKey} eyebrow="Project plugin management" projectID={project?.ID} onSignOut={signOut}>
    <div className="section-heading"><Link href={project ? projectUIURL(project.ID, "plugins") : systemUIURL("projects")}>Back to plugins</Link><button type="button" className="secondary" disabled={loading || busy || !session} onClick={() => { setLoading(true); setMessage(""); void load(session).catch(reason => setMessage(String(reason))).finally(() => setLoading(false)); }}>Refresh installation</button></div>
    <p className="muted">Project: {project?.ID ?? "—"}</p>
    {loading && <AdminProgress label="Loading project installation…" />}
    {message && <p role="alert">{message}</p>}
    {installOpen && project && manifest && <PluginInstallDialog projectID={project.ID} manifest={manifest} onClose={() => setInstallOpen(false)} onInstall={async () => { await installAndActivatePlugin(session, project.ID, manifest); await load(session).catch(reason => setMessage("Plugin installed; refresh failed: " + String(reason))); }} />}
    {pendingAction && project && manifest && <AdminDialog title={pendingAction.label} scope={scope} busy={busy} onClose={() => setPendingAction(undefined)}>
      <p>{pendingAction.capability || (pendingAction.operation === "plugin.readiness" ? "Artifact readiness is shared across projects." : pendingAction.operation === "plugin.enable" ? "Enable the selected plugin and the confirmed required dependencies in this project." : "This action applies only to the selected installation.")}</p>
      {pendingAction.dependencyPlan && <PluginEnableDependencies plan={pendingAction.dependencyPlan} />}
      {pendingAction.operation === "plugin.grant-resource" && <><p className="muted">The first exact resource grant restricts resource-bound operations for this capability to the listed resources. Further grants add exact keys. Other access policies still apply.</p><label>Exact resource key<input disabled={busy || actionDone} value={pendingAction.resource ?? ""} onChange={event => setPendingAction({ ...pendingAction, resource: event.target.value })} /></label></>}
      {pendingAction.operation === "plugin.stop-immediate" && <p role="alert">Stop accepting work immediately; active work may be interrupted.</p>}
      {pendingAction.operation === "plugin.install" && <p>Stage only: no capabilities are granted and no worker is started. Existing routing is unchanged.</p>}
      {pendingAction.operation === "plugin.approve-capability" && <p>Approve this newly requested capability for this exact upgrade version. This is separate from its capability grant.</p>}
      {pendingAction.operation === "plugin.promote" && <p>Switch declared HTTP routing to this exact version. Previously enabled versions remain installed for review or rollback.</p>}
      {pendingAction.operation === "plugin.rollback" && <p>Make this exact version the active version. The host checks compatibility, grants and configuration before accepting it.</p>}
      {pendingAction.operation === "plugin.recover" && <p>Clear the draining or stopped state. Start the worker separately if required.</p>}
      {pendingAction.preview && <p>{pendingAction.preview.Dependents.length ? "Affected dependencies: " + pendingAction.preview.Dependents.join(", ") : "No affected dependencies."}</p>}
      {!actionDone && <button type="button" disabled={busy || (pendingAction.operation === "plugin.enable" && (!pendingAction.dependencyPlan || !!pendingAction.dependencyPlan.blockers.length)) || (pendingAction.operation === "plugin.disable.apply" && !pendingAction.preview) || (pendingAction.operation === "plugin.grant-resource" && !pendingAction.resource?.trim())} onClick={() => void lifecycle()}>Confirm action</button>}
      {busy && <AdminProgress label="Applying change…" />}{actionDone && <p role="status">Action completed</p>}{actionError && <p role="alert">{actionError}</p>}
    </AdminDialog>}
    {manifest && project && review && <>
      <section className="section admin-action-panel"><h2>Installation</h2><p>{manifest.Description}</p><p><code>{manifest.ID}@{manifest.Version}</code> · {!installation ? "Not installed" : installation.Draining ? "Draining" : installation.Enabled ? "Enabled" : "Disabled"} · {installation?.Ready ? "Artifact ready" : "Artifact not ready"} · {installation?.WorkerReady ? "Worker ready" : "Worker not ready"}</p>
        <div className="lifecycle-actions">
          {!installation ? <>{canActivateFresh && <button type="button" onClick={() => setInstallOpen(true)}>Install</button>}{actionButton("plugin.install", "Stage installation")}{!canActivateFresh && <p className="muted">Stage this version and review dependencies and grants before enabling. Existing live versions are not switched automatically.</p>}</> : <>
            {!installation.Ready && actionButton("plugin.readiness", "Mark artifact ready")}
            {installation.Ready && !installation.Enabled && !installation.Draining && <button type="button" disabled={busy || !!review.missingGrants.length || !!enablePlan?.blockers.length || (!!review.otherEnabled.length && !review.promoted)} onClick={() => void openAction("plugin.enable", "Enable")}>Enable</button>}
            {installation.Enabled && installation.Ready && !installation.WorkerReady && actionButton("plugin.worker.start", "Start worker")}
            {installation.Enabled && !installation.Draining && actionButton("plugin.drain", "Drain")}
            {installation.Draining && actionButton("plugin.resume", "Resume")}
            {installation.Enabled && actionButton("plugin.stop-immediate", "Stop immediately", true)}
            {installation.Draining && actionButton("plugin.recover", "Recover")}
            {canPromote && !!review.promoted && !installation.Promoted && actionButton("plugin.rollback", "Roll back to this version", true)}
            {installation.Enabled && actionButton("plugin.disable.apply", "Disable", true)}
            {!installation.Enabled && !installation.Draining && actionButton("plugin.remove", "Uninstall", true)}
            {canPromote && !installation.Promoted && actionButton("plugin.promote", "Promote version")}
          </>}
        </div>
      </section>
      <section className="section admin-action-panel"><h2>Version routing and dependencies</h2><p>{review.promoted ? "Promoted version: " + review.promoted.manifest.Version : "No explicit version promotion. Multiple enabled versions can make HTTP routing ambiguous."}</p>{!!review.otherEnabled.length && !review.promoted && <p role="alert">Promote the currently live version before enabling another version.</p>}<div className="table-wrap"><table><thead><tr><th>Registered version</th><th>Installation state</th><th>Routing</th></tr></thead><tbody>{review.versions.map(item => <tr key={item.manifest.Version}><td><Link href={projectUIURL(project.ID, "plugins", item.manifest.ID + "@" + item.manifest.Version)}>{item.manifest.Version}</Link></td><td>{pluginCatalogState(item.installation)}</td><td>{item.installation?.Promoted ? "Promoted" : "Not promoted"}</td></tr>)}</tbody></table></div>{review.dependencies.length ? <div className="table-wrap"><table><thead><tr><th>Exact dependency</th><th>Kind</th><th>Availability</th></tr></thead><tbody>{review.dependencies.map(dep => <tr key={dep.key}><td>{dep.registered ? <Link href={projectUIURL(project.ID, "plugins", dep.key)}>{dep.key}</Link> : <code>{dep.key}</code>}</td><td>{dep.required ? "Required" : "Optional"}</td><td>{dep.available ? "Admitted by catalog" : dep.selectedOtherVersion ? "Different version promoted" : dep.state}</td></tr>)}</tbody></table></div> : <p className="muted">No declared dependencies.</p>}<p className="muted">Dependency status is a catalog check, not a guarantee of runtime policy authorization.</p>{installation && <><h3>Promotion prerequisites</h3>{!!review.missingGrants.length && <p>Missing grants: {review.missingGrants.join(", ")}</p>}{!!review.missingConfig.length && <p>Missing configuration: {review.missingConfig.join(", ")}</p>}{!!review.incompatibleConfig.length && <p role="alert">Incompatible configuration schema: {review.incompatibleConfig.join(", ")}</p>}{!!review.missingApprovals.length && <p>New capability approvals required: {review.missingApprovals.join(", ")}</p>}{!!review.requiredUnavailable.length && <p role="alert">Required dependencies unavailable for promotion: {review.requiredUnavailable.map(dep => dep.key).join(", ")}</p>}{!!enablePlan?.blockers.length && <p role="alert">Enable blocked: {enablePlan.blockers.join("; ")}</p>}{canPromote && <p>Catalog prerequisites satisfied; host checks again at promotion.</p>}</>}</section>
      {installation ? <>
        <AdminConfigForm key={scope} token={session} projectID={project.ID} pluginID={manifest.ID} version={manifest.Version} schema={manifest.ConfigSchema ?? {}} configuredFields={installation.ConfiguredFields} publicConfig={installation.PublicConfig} maxUploadBytes={installation.MaxUploadBytes} onSaved={() => load(session)} />
        <section className="section admin-action-panel"><h2>Capabilities</h2><p className="muted">Capability grants do not override explicit project access policies. Resource keys are exact, not wildcard patterns.</p><div className="table-wrap"><table><thead><tr><th>Capability</th><th>Scope</th><th>Actions</th></tr></thead><tbody>{manifest.Capabilities.map(capability => {
          const granted = installation.Grants?.includes(capability);
          const resources = installation.ResourceGrants?.[capability] ?? [];
          const supportsResource = manifest.Routes?.some(route => route.Capability === capability && route.ResourceField);
          return <tr key={capability}><td><code>{capability}</code></td><td>{!granted ? "Not granted" : resources.length ? <ul>{resources.map(resource => <li key={resource}><code>{resource}</code></li>)}</ul> : "Capability-wide"}</td><td><div className="lifecycle-actions">
            {granted ? <button type="button" className="danger" disabled={busy} onClick={() => void openAction("plugin.revoke", "Revoke capability and resource grants", capability)}>Revoke</button> : <button type="button" disabled={busy} onClick={() => void openAction("plugin.grant", "Grant capability", capability)}>Grant</button>}
            {review.newlyRequested.includes(capability) && <>{installation.ApprovedCapabilities?.includes(capability) ? <small>Upgrade approved</small> : <button type="button" className="secondary" disabled={busy} onClick={() => void openAction("plugin.approve-capability", "Approve upgrade capability", capability)}>Approve upgrade</button>}</>}
            {supportsResource && <button type="button" className="secondary" disabled={busy} onClick={() => void openAction("plugin.grant-resource", "Grant exact resource", capability)}>Grant resource</button>}
          </div></td></tr>;
        })}</tbody></table></div>{!manifest.Capabilities.length && <p className="muted">No declared capabilities.</p>}</section>
        <section className="section"><h2>Admin pages</h2>{adminContributions(project.ID, { manifests: [manifest], installations: [installation] }).map(item => <p key={item.href}><Link href={item.href}>{item.page.title}</Link></p>)}{manifest.ID === "dashboard" && <p><Link href={systemUIURL("projects")}>Project control plane</Link> · <Link href={projectUIURL(project.ID, "plugins")}>Plugin catalog</Link></p>}{manifest.ID === "functions" && <p><Link href={projectUIURL(project.ID, "functions")}>Function editor and publications</Link></p>}</section>
      </> : <p className="muted">Install this plugin in {project.ID} to edit settings and capabilities.</p>}
    </>}
  </AdminPageFrame>;
}
