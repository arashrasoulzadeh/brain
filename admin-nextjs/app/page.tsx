"use client";
import { FormEvent, useState } from "react";
type Project = { ID: string; Active: boolean };
type PluginManifest = { ID: string; Title: string; Description: string; Version: string; Capabilities: string[] };
type PluginStatus = { PluginID: string; Version: string; Enabled: boolean; Ready: boolean; Draining: boolean };
type AuditRecord = { source: string; action: string; actor_id?: string; project_id?: string; plugin_id?: string; version?: string; success: boolean; detail?: string };
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
  },
};
export default function AdminPage() {
  const [language, setLanguage] = useState<Language>("en");
  const [session, setSession] = useState("");
  const [projects, setProjects] = useState<Project[]>([]);
  const [manifests, setManifests] = useState<PluginManifest[]>([]);
  const [installations, setInstallations] = useState<PluginStatus[]>([]);
  const [audit, setAudit] = useState<AuditRecord[]>([]);
  const [username, setUsername] = useState("");
  const [secret, setSecret] = useState("");
  const [projectID, setProjectID] = useState("");
  const [error, setError] = useState("");
  const text = copy[language];
  async function loadProjects(token: string) {
    const response = await fetch("/api/admin/projects", {
      headers: { authorization: `Bearer ${token}` },
    });
    if (!response.ok) throw new Error("Could not load projects");
    setProjects((await response.json()) as Project[]);
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
      await loadProjects(token);
      await loadPlugins(token);
      await loadAudit(token);
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
    const project = projects.find((candidate) => candidate.Active);
    if (!project) { setError("Create an active project first"); return; }
    try {
      const response = await fetch("/api/admin/actions", { method: "POST", headers: { authorization: `Bearer ${session}`, "content-type": "application/json" }, body: JSON.stringify({ operation: "plugin.install", project_id: project.ID, plugin_id: manifest.ID, version: manifest.Version }) });
      if (!response.ok) throw new Error(await response.text());
      await loadPlugins(session);
    } catch (reason) { setError(reason instanceof Error ? reason.message : "Could not install plugin"); }
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
                <small>{project.Active ? text.active : text.inactive}</small>
              </div>
            ))
          ) : (
            <p className="muted">{text.empty}</p>
          )}
        </div>
        <section className="section">
          <div className="section-heading"><div><p className="eyebrow">{text.plugins}</p><h2>{text.registered}</h2></div><span className="count">{manifests.length}</span></div>
          <div className="plugin-grid">{manifests.length ? manifests.map((manifest) => { const installed = installations.some((item) => item.PluginID === manifest.ID && item.Version === manifest.Version); return <article className="plugin" key={`${manifest.ID}@${manifest.Version}`}><div><h3>{manifest.Title || manifest.ID}</h3><p className="muted">{manifest.Description}</p></div><small>{manifest.ID}@{manifest.Version}</small><div className="capabilities">{manifest.Capabilities?.slice(0, 3).map((capability) => <span key={capability}>{capability}</span>)}</div><button className="plugin-action" disabled={installed} onClick={() => installPlugin(manifest)}>{installed ? text.installedLabel : text.install}</button></article>; }) : <p className="muted">{text.empty}</p>}</div>
        </section>
        <section className="section"><div className="section-heading"><div><p className="eyebrow">{text.plugins}</p><h2>{text.installed}</h2></div><span className="count">{installations.length}</span></div><div>{installations.map((installation) => <div className="project" key={`${installation.PluginID}@${installation.Version}`}><span>{installation.PluginID}@{installation.Version}</span><small>{installation.Enabled ? text.enabled : text.disabled} · {installation.Ready ? text.ready : text.notReady}</small></div>)}</div></section>
        <section className="section"><div className="section-heading"><div><p className="eyebrow">{text.audit}</p><h2>{text.audit}</h2></div><span className="count">{audit.length}</span></div><div className="audit-list">{audit.length ? audit.slice().reverse().slice(0, 20).map((record, index) => <div className="audit" key={`${record.action}-${index}`}><span className={record.success ? "success" : "failure"}>{record.success ? "✓" : "!"}</span><div><strong>{record.action}</strong><small>{record.source}{record.plugin_id ? ` · ${record.plugin_id}@${record.version}` : ""}{record.project_id ? ` · ${record.project_id}` : ""}</small></div></div>) : <p className="muted">{text.noAudit}</p>}</div></section>
      </section>
    </main>
  );
}
