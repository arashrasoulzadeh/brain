"use client";
import { FormEvent, useState } from "react";
type Project = { ID: string; Active: boolean };
type PluginManifest = { ID: string; Title: string; Description: string; Version: string; Capabilities: string[] };
type PluginStatus = { PluginID: string; Version: string; Enabled: boolean; Ready: boolean; Draining: boolean };
type AuditRecord = { source: string; action: string; actor_id?: string; project_id?: string; plugin_id?: string; version?: string; success: boolean; detail?: string };
type SystemManifest = PluginManifest & { Routes: { Operation: string; Capability: string }[]; ExtensionPoints: string[] };
type Language = "en" | "fa";
const copy = {
  en: {
    eyebrow: "BRAIN CONTROL PLANE",
    welcome: "Welcome back",
    subtitle: "Securely manage projects and plugins.",
    username: "Operator username",
    secret: "Secret",
    signIn: "Sign in",
    projects: "Projects",
    signOut: "Sign out",
    newProject: "New project ID",
    create: "Create project",
    empty: "No projects yet.",
    active: "Active",
    inactive: "Inactive",
    plugins: "Plugins",
    registered: "Registered plugins",
    installed: "Installed versions",
    ready: "Ready",
    notReady: "Not ready",
    enabled: "Enabled",
    disabled: "Disabled",
    invalid: "Invalid operator credentials",
    install: "Install",
    installedLabel: "Installed",
    audit: "Audit log",
    noAudit: "No audit events yet.",
    system: "System capabilities",
    routes: "routes",
    extensions: "extension points",
    navigation: "Admin navigation",
    project: "Project",
    uninstall: "Uninstall",
    enable: "Enable",
    grant: "Grant first capability",
    remove: "Remove",
    markReady: "Mark ready",
    promote: "Promote",
    drain: "Drain",
    recover: "Recover",
    rollback: "Rollback",
  },
  fa: {
    eyebrow: "پنل مدیریت برین",
    welcome: "خوش آمدید",
    subtitle: "پروژه‌ها و افزونه‌ها را با امنیت مدیریت کنید.",
    username: "نام کاربری مدیر",
    secret: "رمز عبور",
    signIn: "ورود",
    projects: "پروژه‌ها",
    signOut: "خروج",
    newProject: "شناسه پروژه جدید",
    create: "ساخت پروژه",
    empty: "هنوز پروژه‌ای وجود ندارد.",
    active: "فعال",
    inactive: "غیرفعال",
    plugins: "افزونه‌ها",
    registered: "افزونه‌های ثبت‌شده",
    installed: "نسخه‌های نصب‌شده",
    ready: "آماده",
    notReady: "آماده نیست",
    enabled: "فعال",
    disabled: "غیرفعال",
    invalid: "اطلاعات ورود نادرست است",
    install: "نصب",
    installedLabel: "نصب‌شده",
    audit: "گزارش رویدادها",
    noAudit: "هنوز رویدادی ثبت نشده است.",
    system: "قابلیت‌های سیستم",
    routes: "مسیر",
    extensions: "نقاط توسعه",
    navigation: "ناوبری مدیریت",
    project: "پروژه",
    uninstall: "حذف پروژه",
    enable: "فعال‌سازی",
    grant: "اعطای اولین دسترسی",
    remove: "حذف",
    markReady: "اعلام آمادگی",
    promote: "ارتقا",
    drain: "تخلیه",
    recover: "بازیابی",
    rollback: "بازگشت",
  },
};
export default function AdminPage() {
  const [language, setLanguage] = useState<Language>("en");
  const [session, setSession] = useState("");
  const [projects, setProjects] = useState<Project[]>([]);
  const [manifests, setManifests] = useState<PluginManifest[]>([]);
  const [installations, setInstallations] = useState<PluginStatus[]>([]);
  const [audit, setAudit] = useState<AuditRecord[]>([]);
  const [system, setSystem] = useState<SystemManifest[]>([]);
  const [username, setUsername] = useState("");
  const [secret, setSecret] = useState("");
  const [projectID, setProjectID] = useState("");
  const [selectedProjectID, setSelectedProjectID] = useState("");
  const [error, setError] = useState("");
  const text = copy[language];
  async function loadProjects(token: string): Promise<Project[]> {
    const response = await fetch("/api/admin/projects", {
      headers: { authorization: `Bearer ${token}` },
    });
    if (!response.ok) throw new Error("Could not load projects");
    const loadedProjects = (await response.json()) as Project[];
    setProjects(loadedProjects);
    return loadedProjects;
  }
  async function loadPlugins(token: string) {
    const response = await fetch("/api/admin/plugins", { headers: { authorization: `Bearer ${token}` } });
    if (!response.ok) throw new Error("Could not load plugins");
    const result = (await response.json()) as { manifests: PluginManifest[]; installations: PluginStatus[] };
    setManifests(result.manifests);
    setInstallations(result.installations);
  }
  async function loadAudit(token: string) {
    const response = await fetch("/api/admin/audit", { headers: { authorization: `Bearer ${token}` } });
    if (!response.ok) throw new Error("Could not load audit log");
    setAudit((await response.json()) as AuditRecord[]);
  }
  async function loadSystem(token: string) {
    const response = await fetch("/api/admin/system", { headers: { authorization: `Bearer ${token}` } });
    if (!response.ok) throw new Error("Could not load system capabilities");
    setSystem((await response.json()) as SystemManifest[]);
  }
  async function signIn(event: FormEvent) {
    event.preventDefault();
    setError("");
    try {
      const response = await fetch("/api/admin/session", {
        method: "POST",
        headers: { "content-type": "application/json" },
        body: JSON.stringify({ username, secret }),
      });
      if (!response.ok) throw new Error(text.invalid);
      const token = ((await response.json()) as { session: string }).session;
      setSession(token);
      const loadedProjects = await loadProjects(token);
      setSelectedProjectID(loadedProjects.find((project) => project.Active)?.ID ?? "");
      await loadPlugins(token);
      await loadAudit(token);
      await loadSystem(token);
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Sign in failed");
    }
  }
  async function createProject(event: FormEvent) {
    event.preventDefault();
    setError("");
    try {
      const response = await fetch("/api/admin/actions", {
        method: "POST",
        headers: { authorization: `Bearer ${session}`, "content-type": "application/json" },
        body: JSON.stringify({ operation: "project.create", project_id: projectID }),
      });
      if (!response.ok) throw new Error(await response.text());
      setProjectID("");
      await loadProjects(session);
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Could not create project");
    }
  }
  async function installPlugin(manifest: PluginManifest) {
    setError("");
    const project = projects.find((candidate) => candidate.ID === selectedProjectID && candidate.Active);
    if (!project) { setError("Create an active project first"); return; }
    try {
      const response = await fetch("/api/admin/actions", { method: "POST", headers: { authorization: `Bearer ${session}`, "content-type": "application/json" }, body: JSON.stringify({ operation: "plugin.install", project_id: project.ID, plugin_id: manifest.ID, version: manifest.Version }) });
      if (!response.ok) throw new Error(await response.text());
      await loadPlugins(session);
    } catch (reason) { setError(reason instanceof Error ? reason.message : "Could not install plugin"); }
  }
  async function uninstallProject(project: Project) {
    if (!window.confirm(`${text.uninstall} ${project.ID}?`)) return;
    setError("");
    const response = await fetch("/api/admin/actions", { method: "POST", headers: { authorization: `Bearer ${session}`, "content-type": "application/json" }, body: JSON.stringify({ operation: "project.uninstall", project_id: project.ID, confirmed: true }) });
    if (!response.ok) { setError(await response.text()); return; }
    const loadedProjects = await loadProjects(session);
    setSelectedProjectID(loadedProjects.find((candidate) => candidate.Active)?.ID ?? "");
    await loadPlugins(session); await loadAudit(session);
  }
  async function runPluginAction(operation: string, installation: PluginStatus) {
    setError("");
    const project = projects.find((candidate) => candidate.Active);
    if (!project) { setError("Create an active project first"); return; }
    const manifest = manifests.find((candidate) => candidate.ID === installation.PluginID && candidate.Version === installation.Version);
    const capability = manifest?.Capabilities?.[0] ?? "";
    if ((operation === "plugin.remove" || operation === "plugin.rollback" || operation === "plugin.stop-immediate") && !window.confirm(`${operation} ${installation.PluginID}@${installation.Version}?`)) return;
    const response = await fetch("/api/admin/actions", { method: "POST", headers: { authorization: `Bearer ${session}`, "content-type": "application/json" }, body: JSON.stringify({ operation, project_id: project.ID, plugin_id: installation.PluginID, version: installation.Version, capability, ready: operation === "plugin.readiness", confirmed: ["plugin.remove", "plugin.rollback", "plugin.stop-immediate"].includes(operation) }) });
    if (!response.ok) { setError(await response.text()); return; }
    await loadPlugins(session); await loadAudit(session);
  }
  const switcher = (
    <button
      className="language"
      type="button"
      onClick={() => setLanguage(language === "en" ? "fa" : "en")}
    >
      {language === "en" ? "فارسی" : "English"}
    </button>
  );
  if (!session)
    return (
      <main className="shell" dir={language === "fa" ? "rtl" : "ltr"}>
        <section className="card login">
          {switcher}
          <p className="eyebrow">{text.eyebrow}</p>
          <h1>{text.welcome}</h1>
          <p className="muted">{text.subtitle}</p>
          <form onSubmit={signIn}>
            <label>
              {text.username}
              <input
                value={username}
                onChange={(event) => setUsername(event.target.value)}
                required
              />
            </label>
            <label>
              {text.secret}
              <input
                type="password"
                value={secret}
                onChange={(event) => setSecret(event.target.value)}
                required
              />
            </label>
            <button>{text.signIn}</button>
            {error && <p className="error">{error}</p>}
          </form>
        </section>
      </main>
    );
  return (
    <main className="shell" dir={language === "fa" ? "rtl" : "ltr"}>
      <section className="card">
        <div className="top">
          <div>
            <p className="eyebrow">{text.eyebrow}</p>
            <h1>{text.projects}</h1>
          </div>
          <div>
            {switcher}
            <button className="secondary" onClick={() => setSession("")}>
              {text.signOut}
            </button>
          </div>
        </div>
        <nav className="admin-nav" aria-label={text.navigation}><a href="#projects">{text.projects}</a><a href="#plugins">{text.plugins}</a><a href="#system">{text.system}</a><a href="#audit">{text.audit}</a></nav>
        <div className="project-selector"><label htmlFor="selected-project">{text.project}</label><select id="selected-project" value={selectedProjectID} onChange={(event) => setSelectedProjectID(event.target.value)}><option value="">—</option>{projects.map((project) => <option key={project.ID} value={project.ID} disabled={!project.Active}>{project.ID}{project.Active ? "" : " (inactive)"}</option>)}</select></div>
        <form className="create" onSubmit={createProject}>
          <input
            placeholder={text.newProject}
            value={projectID}
            onChange={(event) => setProjectID(event.target.value)}
            required
          />
          <button>{text.create}</button>
        </form>
        {error && <p className="error">{error}</p>}
        <div>
          {projects.length ? (
            projects.map((project) => (
              <div className="project" key={project.ID}>
                <span>{project.ID}</span>
                <span className="project-actions"><small>{project.Active ? text.active : text.inactive}</small><button className="danger" onClick={() => uninstallProject(project)}>{text.uninstall}</button></span>
              </div>
            ))
          ) : (
            <p className="muted">{text.empty}</p>
          )}
        </div>
        <section className="section" id="projects">
          <div className="section-heading"><div><p className="eyebrow">{text.plugins}</p><h2>{text.registered}</h2></div><span className="count">{manifests.length}</span></div>
          <div className="plugin-grid">{manifests.length ? manifests.map((manifest) => { const installed = installations.some((item) => item.PluginID === manifest.ID && item.Version === manifest.Version); return <article className="plugin" key={`${manifest.ID}@${manifest.Version}`}><div><h3>{manifest.Title || manifest.ID}</h3><p className="muted">{manifest.Description}</p></div><small>{manifest.ID}@{manifest.Version}</small><div className="capabilities">{manifest.Capabilities?.slice(0, 3).map((capability) => <span key={capability}>{capability}</span>)}</div><button className="plugin-action" disabled={installed} onClick={() => installPlugin(manifest)}>{installed ? text.installedLabel : text.install}</button></article>; }) : <p className="muted">{text.empty}</p>}</div>
        </section>
        <section className="section" id="plugins"><div className="section-heading"><div><p className="eyebrow">{text.plugins}</p><h2>{text.installed}</h2></div><span className="count">{installations.length}</span></div><div>{installations.map((installation) => <div className="project lifecycle-row" key={`${installation.PluginID}@${installation.Version}`}><span>{installation.PluginID}@{installation.Version}<small>{installation.Enabled ? text.enabled : text.disabled} · {installation.Ready ? text.ready : text.notReady}</small></span><span className="lifecycle-actions">{!installation.Ready && <button onClick={() => runPluginAction("plugin.readiness", installation)}>{text.markReady}</button>}{!installation.Enabled && <button onClick={() => runPluginAction("plugin.enable", installation)}>{text.enable}</button>}{!installation.Enabled && <button onClick={() => runPluginAction("plugin.grant", installation)}>{text.grant}</button>}{installation.Enabled && <button onClick={() => runPluginAction("plugin.promote", installation)}>{text.promote}</button>}<button onClick={() => runPluginAction("plugin.drain", installation)}>{text.drain}</button><button onClick={() => runPluginAction("plugin.recover", installation)}>{text.recover}</button><button onClick={() => runPluginAction("plugin.rollback", installation)}>{text.rollback}</button><button className="danger" onClick={() => runPluginAction("plugin.remove", installation)}>{text.remove}</button></span></div>)}</div></section>
        <section className="section" id="audit"><div className="section-heading"><div><p className="eyebrow">{text.audit}</p><h2>{text.audit}</h2></div><span className="count">{audit.length}</span></div><div className="audit-list">{audit.length ? audit.slice().reverse().slice(0, 20).map((record, index) => <div className="audit" key={`${record.action}-${index}`}><span className={record.success ? "success" : "failure"}>{record.success ? "✓" : "!"}</span><div><strong>{record.action}</strong><small>{record.source}{record.plugin_id ? ` · ${record.plugin_id}@${record.version}` : ""}{record.project_id ? ` · ${record.project_id}` : ""}</small></div></div>) : <p className="muted">{text.noAudit}</p>}</div></section>
        <section className="section" id="system"><div className="section-heading"><div><p className="eyebrow">{text.system}</p><h2>{text.system}</h2></div><span className="count">{system.length}</span></div><div className="plugin-grid">{system.map((manifest) => <article className="plugin" key={`system-${manifest.ID}@${manifest.Version}`}><h3>{manifest.Title || manifest.ID}</h3><p className="muted">{manifest.Description}</p><small>{manifest.Routes?.length ?? 0} {text.routes} · {manifest.ExtensionPoints?.length ?? 0} {text.extensions}</small><div className="capabilities">{manifest.Capabilities?.map((capability) => <span key={capability}>{capability}</span>)}</div><div className="route-list">{manifest.Routes?.slice(0, 8).map((route) => <div key={route.Operation}><code>{route.Operation}</code><small>{route.Capability}</small></div>)}</div></article>)}</div></section>
      </section>
    </main>
  );
}
