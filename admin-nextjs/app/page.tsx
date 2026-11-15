"use client";
import { FormEvent, useEffect, useState } from "react";
import { usePathname } from "next/navigation";
type Project = { ID: string; Active: boolean };
type PluginManifest = { ID: string; Title: string; Description: string; Version: string; Capabilities: string[]; ConfigSchema?: Record<string, string> };
type PluginStatus = { ProjectID: string; PluginID: string; Version: string; Enabled: boolean; Ready: boolean; Draining: boolean };
type AuditRecord = { source: string; action: string; actor_id?: string; project_id?: string; plugin_id?: string; version?: string; success: boolean; detail?: string };
type SystemManifest = PluginManifest & { Routes: { Operation: string; Capability: string }[]; ExtensionPoints: string[]; SourceFiles: string[]; Imports: string[]; PluginImports: string[]; Dependencies: string[] };
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
    activate: "Activate",
    deactivate: "Deactivate",
    enable: "Enable",
    grant: "Grant first capability",
    grantCapability: "Grant",
    revokeCapability: "Revoke",
    disable: "Disable",
    resume: "Resume",
    stop: "Stop",
    files: "files",
    imports: "imports",
    dependencies: "dependencies",
    search: "Search plugins",
    searchAudit: "Search audit history",
    database: "Database",
    functions: "Functions",
    workflows: "Workflows",
    liveSessions: "Live sessions",
    routesPage: "Routes",
    monitoring: "Monitoring",
    playground: "Playground",
    selectRoute: "Select a route",
    payload: "JSON payload",
    previewRequest: "Preview request",
    test: "Test",
    send: "Send",
    remove: "Remove",
    markReady: "Mark ready",
    promote: "Promote",
    drain: "Drain",
    recover: "Recover",
    rollback: "Rollback",
    configure: "Save configuration",
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
    activate: "فعال‌سازی",
    deactivate: "غیرفعال‌سازی",
    enable: "فعال‌سازی",
    grant: "اعطای اولین دسترسی",
    grantCapability: "اعطا",
    revokeCapability: "لغو دسترسی",
    disable: "غیرفعال‌سازی",
    resume: "ادامه فعالیت",
    stop: "توقف",
    files: "فایل",
    imports: "واردسازی",
    dependencies: "وابستگی",
    search: "جستجوی افزونه‌ها",
    searchAudit: "جستجوی گزارش‌ها",
    database: "پایگاه داده",
    functions: "توابع",
    workflows: "گردش‌کارها",
    liveSessions: "جلسه‌های زنده",
    routesPage: "مسیرها",
    monitoring: "پایش",
    playground: "محیط آزمایش",
    selectRoute: "انتخاب مسیر",
    payload: "داده JSON",
    previewRequest: "پیش‌نمایش درخواست",
    test: "آزمایش",
    send: "ارسال",
    remove: "حذف",
    markReady: "اعلام آمادگی",
    promote: "ارتقا",
    drain: "تخلیه",
    recover: "بازیابی",
    rollback: "بازگشت",
    configure: "ذخیره تنظیمات",
  },
};
export default function AdminPage() {
  const pathname = usePathname();
  const activePage = pathname.split("/")[1] || "projects";
  const [language, setLanguage] = useState<Language>("en");
  const [session, setSession] = useState("");
  const [hydratingSession, setHydratingSession] = useState(true);
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
  const [configuration, setConfiguration] = useState<Record<string, Record<string, string | boolean>>>({});
  const [pluginFilter, setPluginFilter] = useState("");
  const [auditFilter, setAuditFilter] = useState("");
  const [playgroundRoute, setPlaygroundRoute] = useState("");
  const [playgroundPayload, setPlaygroundPayload] = useState("{}");
  const [playgroundMessage, setPlaygroundMessage] = useState("");
  const [functionName, setFunctionName] = useState("");
  const [functionSource, setFunctionSource] = useState("export default async function handler(input) {\n  return input;\n}\n");
  const [functionMessage, setFunctionMessage] = useState("");
  const text = copy[language];
  const pageTitle = ({ projects: text.projects, plugins: text.plugins, system: text.system, audit: text.audit, database: text.database, functions: text.functions, workflows: text.workflows, files: text.files, "live-sessions": text.liveSessions, routes: text.routesPage, monitoring: text.monitoring } as Record<string, string>)[activePage] ?? text.projects;
  useEffect(() => {
    if (!session) return;
    const refresh = () => Promise.all([loadProjects(session), loadPlugins(session, selectedProjectID), loadAudit(session), loadSystem(session)]).catch(() => undefined);
    const timer = window.setInterval(refresh, 10000);
    return () => window.clearInterval(timer);
  }, [session, selectedProjectID]);
  useEffect(() => {
    const savedSession = window.sessionStorage.getItem("brain-admin-session");
    if (!savedSession) { setHydratingSession(false); return; }
    setSession(savedSession);
    loadProjects(savedSession).then((loadedProjects) => {
      const activeProjectID = loadedProjects.find((project) => project.Active)?.ID ?? "";
      setSelectedProjectID(activeProjectID);
      return Promise.all([loadPlugins(savedSession, activeProjectID), loadAudit(savedSession), loadSystem(savedSession)]);
    }).catch(() => {
      window.sessionStorage.removeItem("brain-admin-session");
      setSession("");
      setError("Your operator session expired. Please sign in again.");
    }).finally(() => setHydratingSession(false));
  }, []);
  async function loadProjects(token: string): Promise<Project[]> {
    const response = await fetch("/api/admin/projects", {
      headers: { authorization: `Bearer ${token}` },
    });
    if (!response.ok) throw new Error("Could not load projects");
    const loadedProjects = (await response.json()) as Project[];
    setProjects(loadedProjects);
    return loadedProjects;
  }
  async function loadPlugins(token: string, projectID = selectedProjectID) {
    const query = projectID ? `?project_id=${encodeURIComponent(projectID)}` : "";
    const response = await fetch(`/api/admin/plugins${query}`, { headers: { authorization: `Bearer ${token}` } });
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
      window.sessionStorage.setItem("brain-admin-session", token);
      const loadedProjects = await loadProjects(token);
      setSelectedProjectID(loadedProjects.find((project) => project.Active)?.ID ?? "");
      await loadPlugins(token, loadedProjects.find((project) => project.Active)?.ID ?? "");
      await loadAudit(token);
      await loadSystem(token);
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Sign in failed");
    }
  }
  async function signOut() {
    await fetch("/api/admin/session", { method: "DELETE", headers: { authorization: `Bearer ${session}` } }).catch(() => undefined);
    window.sessionStorage.removeItem("brain-admin-session");
    setSession("");
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
  async function setProjectActive(project: Project, active: boolean) {
    if (!active && !window.confirm(`${text.deactivate} ${project.ID}?`)) return;
    const response = await fetch("/api/admin/actions", { method: "POST", headers: { authorization: `Bearer ${session}`, "content-type": "application/json" }, body: JSON.stringify({ operation: active ? "project.activate" : "project.deactivate", project_id: project.ID, confirmed: !active }) });
    if (!response.ok) { setError(await response.text()); return; }
    const loadedProjects = await loadProjects(session);
    if (selectedProjectID === project.ID && !active) setSelectedProjectID(loadedProjects.find((candidate) => candidate.Active)?.ID ?? "");
    await loadPlugins(session, active ? project.ID : ""); await loadAudit(session);
  }
  function testPlayground() {
    try { JSON.parse(playgroundPayload || "{}"); setPlaygroundMessage("Valid JSON payload"); } catch { setPlaygroundMessage("Invalid JSON payload"); }
  }
  function sendPlayground() {
    testPlayground();
    if (!playgroundRoute) { setPlaygroundMessage("Select a route first"); return; }
    setPlaygroundMessage("Request prepared for the selected project; host execution requires an installed, ready plugin.");
  }
  async function installPlugin(manifest: PluginManifest) {
    setError("");
    const project = projects.find((candidate) => candidate.ID === selectedProjectID && candidate.Active);
    if (!project) { setError("Create an active project first"); return; }
    try {
      const response = await fetch("/api/admin/actions", { method: "POST", headers: { authorization: `Bearer ${session}`, "content-type": "application/json" }, body: JSON.stringify({ operation: "plugin.install", project_id: project.ID, plugin_id: manifest.ID, version: manifest.Version }) });
      if (!response.ok) throw new Error(await response.text());
      await loadPlugins(session, selectedProjectID);
    } catch (reason) { setError(reason instanceof Error ? reason.message : "Could not install plugin"); }
  }
  async function submitFunction(operation: "functions.test" | "functions.create") {
    if (!selectedProjectID || !functionName.trim() || !functionSource.trim()) { setFunctionMessage("Select an active project, name the function, and provide source."); return; }
    const response = await fetch("/api/admin/functions", { method: "POST", headers: { authorization: `Bearer ${session}`, "content-type": "application/json" }, body: JSON.stringify({ operation, project_id: selectedProjectID, name: functionName, source: functionSource }) });
    setFunctionMessage(response.ok ? operation === "functions.test" ? "Function source accepted for testing." : "Function draft saved." : await response.text());
  }
  async function uninstallProject(project: Project) {
    if (!window.confirm(`${text.uninstall} ${project.ID}?`)) return;
    setError("");
    const response = await fetch("/api/admin/actions", { method: "POST", headers: { authorization: `Bearer ${session}`, "content-type": "application/json" }, body: JSON.stringify({ operation: "project.uninstall", project_id: project.ID, confirmed: true }) });
    if (!response.ok) { setError(await response.text()); return; }
    const loadedProjects = await loadProjects(session);
    setSelectedProjectID(loadedProjects.find((candidate) => candidate.Active)?.ID ?? "");
    await loadPlugins(session, loadedProjects.find((candidate) => candidate.Active)?.ID ?? ""); await loadAudit(session);
  }
  async function runPluginAction(operation: string, installation: PluginStatus, requestedCapability?: string) {
    setError("");
    const project = projects.find((candidate) => candidate.ID === selectedProjectID && candidate.Active);
    if (!project) { setError("Create an active project first"); return; }
    const manifest = manifests.find((candidate) => candidate.ID === installation.PluginID && candidate.Version === installation.Version);
    const capability = requestedCapability ?? manifest?.Capabilities?.[0] ?? "";
    if ((operation === "plugin.remove" || operation === "plugin.revoke" || operation === "plugin.rollback" || operation === "plugin.stop" || operation === "plugin.stop-immediate") && !window.confirm(`${operation} ${installation.PluginID}@${installation.Version}?`)) return;
    const response = await fetch("/api/admin/actions", { method: "POST", headers: { authorization: `Bearer ${session}`, "content-type": "application/json" }, body: JSON.stringify({ operation, project_id: project.ID, plugin_id: installation.PluginID, version: installation.Version, capability, ready: operation === "plugin.readiness", confirmed: ["plugin.remove", "plugin.revoke", "plugin.rollback", "plugin.stop", "plugin.stop-immediate"].includes(operation) }) });
    if (!response.ok) { setError(await response.text()); return; }
    await loadPlugins(session, selectedProjectID); await loadAudit(session);
  }
  async function configurePlugin(installation: PluginStatus) {
    if (!selectedProjectID) { setError("Select an active project first"); return; }
    const key = `${installation.PluginID}@${installation.Version}`;
    const manifest = manifests.find((candidate) => candidate.ID === installation.PluginID && candidate.Version === installation.Version);
    const rawValues = configuration[key] ?? {};
    const values = Object.fromEntries(Object.entries(rawValues).map(([name, value]) => [name, manifest?.ConfigSchema?.[name] === "int" ? Number(value) : value]));
    const response = await fetch("/api/admin/actions", { method: "POST", headers: { authorization: `Bearer ${session}`, "content-type": "application/json" }, body: JSON.stringify({ operation: "plugin.configure", project_id: selectedProjectID, plugin_id: installation.PluginID, version: installation.Version, values }) });
    if (!response.ok) { setError(await response.text()); return; }
    await loadAudit(session);
  }
  async function disablePlugin(installation: PluginStatus) {
    if (!selectedProjectID) { setError("Select an active project first"); return; }
    const previewResponse = await fetch("/api/admin/actions", { method: "POST", headers: { authorization: `Bearer ${session}`, "content-type": "application/json" }, body: JSON.stringify({ operation: "plugin.disable.preview", project_id: selectedProjectID, plugin_id: installation.PluginID, version: installation.Version }) });
    if (!previewResponse.ok) { setError(await previewResponse.text()); return; }
    const preview = (await previewResponse.json()) as { Revision: number; Dependents: string[] };
    const dependentText = preview.Dependents?.length ? ` Dependencies affected: ${preview.Dependents.join(", ")}.` : "";
    if (!window.confirm(`${text.disable} ${installation.PluginID}@${installation.Version}?${dependentText}`)) return;
    const applyResponse = await fetch("/api/admin/actions", { method: "POST", headers: { authorization: `Bearer ${session}`, "content-type": "application/json" }, body: JSON.stringify({ operation: "plugin.disable.apply", project_id: selectedProjectID, plugin_id: installation.PluginID, version: installation.Version, revision: preview.Revision, dependents: preview.Dependents, confirmed: true }) });
    if (!applyResponse.ok) { setError(await applyResponse.text()); return; }
    await loadPlugins(session, selectedProjectID); await loadAudit(session);
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
  if (hydratingSession)
    return <main className="shell" dir={language === "fa" ? "rtl" : "ltr"}><section className="card login"><p className="eyebrow">{text.eyebrow}</p><p className="muted">Loading operator session…</p></section></main>;
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
      <div className="workspace">
        <aside className="sidebar">
          <p className="eyebrow">{text.eyebrow}</p>
          <h2>{text.projects}</h2>
          <nav className="sidebar-nav" aria-label={text.navigation}>
            <a className={activePage === "projects" ? "active" : ""} href="/projects">{text.projects}</a>
            {manifests.length > 0 && <a className={activePage === "plugins" ? "active" : ""} href="/plugins">{text.plugins}</a>}
            {system.length > 0 && <a className={activePage === "system" ? "active" : ""} href="/system">{text.system}</a>}
            {audit.length > 0 && <a className={activePage === "audit" ? "active" : ""} href="/audit">{text.audit}</a>}
            {system.some((manifest) => manifest.ID === "tables") && <a className={activePage === "database" ? "active" : ""} href="/database">{text.database}</a>}
            {system.some((manifest) => manifest.ID === "functions") && <a className={activePage === "functions" ? "active" : ""} href="/functions">{text.functions}</a>}
            {system.some((manifest) => manifest.ID === "functions") && <a className={activePage === "workflows" ? "active" : ""} href="/workflows">{text.workflows}</a>}
            {system.some((manifest) => manifest.ID === "files") && <a className={activePage === "files" ? "active" : ""} href="/files">{text.files}</a>}
            {system.some((manifest) => ["rooms", "matchmaking"].includes(manifest.ID)) && <a className={activePage === "live-sessions" ? "active" : ""} href="/live-sessions">{text.liveSessions}</a>}
            {system.length > 0 && <a className={activePage === "routes" ? "active" : ""} href="/routes">{text.routesPage}</a>}
            {system.length > 0 && <a className={activePage === "monitoring" ? "active" : ""} href="/monitoring">{text.monitoring}</a>}
          </nav>
          <label className="sidebar-project" htmlFor="sidebar-project">{text.project}<select id="sidebar-project" value={selectedProjectID} onChange={(event) => { const nextProjectID = event.target.value; setSelectedProjectID(nextProjectID); void loadPlugins(session, nextProjectID); }}><option value="">—</option>{projects.map((project) => <option key={project.ID} value={project.ID} disabled={!project.Active}>{project.ID}</option>)}</select></label>
          <div className="sidebar-actions">{switcher}<button className="secondary" onClick={() => void signOut()}>{text.signOut}</button></div>
        </aside>
        <section className="card">
        <div className="top">
          <div>
            <p className="eyebrow">{text.eyebrow}</p>
            <h1>{pageTitle}</h1>
          </div>
        </div>
        <div className={activePage === "projects" ? "project-selector" : "project-selector page-hidden"}><label htmlFor="selected-project">{text.project}</label><select id="selected-project" value={selectedProjectID} onChange={(event) => { const nextProjectID = event.target.value; setSelectedProjectID(nextProjectID); void loadPlugins(session, nextProjectID); }}><option value="">—</option>{projects.map((project) => <option key={project.ID} value={project.ID} disabled={!project.Active}>{project.ID}{project.Active ? "" : " (inactive)"}</option>)}</select></div>
        <form className={activePage === "projects" ? "create" : "create page-hidden"} onSubmit={createProject}>
          <input
            placeholder={text.newProject}
            value={projectID}
            onChange={(event) => setProjectID(event.target.value)}
            required
          />
          <button>{text.create}</button>
        </form>
        {error && <p className="error">{error}</p>}
        <div className={activePage === "projects" ? "" : "page-hidden"}>
          {projects.length ? <div className="project-table"><div className="project-table-head"><span>{text.project}</span><span>{text.active}</span><span>{text.navigation}</span></div>{projects.map((project) => <div className="project-table-row" key={project.ID}><strong>{project.ID}</strong><span className={project.Active ? "success" : "failure"}>{project.Active ? text.active : text.inactive}</span><span className="project-actions">{project.Active ? <button className="secondary" onClick={() => void setProjectActive(project, false)}>{text.deactivate}</button> : <button onClick={() => void setProjectActive(project, true)}>{text.activate}</button>}{project.Active && <button className="danger" onClick={() => uninstallProject(project)}>{text.uninstall}</button>}</span></div>)}</div> : <p className="muted">{text.empty}</p>}
        </div>
        <section className="section page-hidden" id="projects">
          <div className="section-heading"><div><p className="eyebrow">{text.plugins}</p><h2>{text.registered}</h2></div><span className="count">{manifests.length}</span></div>
          <input className="plugin-search" placeholder={text.search} value={pluginFilter} onChange={(event) => setPluginFilter(event.target.value)} />
          <div className="plugin-grid">{manifests.filter((manifest) => `${manifest.ID} ${manifest.Title} ${manifest.Description}`.toLowerCase().includes(pluginFilter.toLowerCase())).length ? manifests.filter((manifest) => `${manifest.ID} ${manifest.Title} ${manifest.Description}`.toLowerCase().includes(pluginFilter.toLowerCase())).map((manifest) => { const installation = installations.find((item) => item.PluginID === manifest.ID && item.Version === manifest.Version); const installed = Boolean(installation); return <article className="plugin" key={`${manifest.ID}@${manifest.Version}`}><div><h3><a href={`/plugins/${encodeURIComponent(`${manifest.ID}@${manifest.Version}`)}`}>{manifest.Title || manifest.ID}</a></h3><p className="muted">{manifest.Description}</p></div><small>{manifest.ID}@{manifest.Version}</small><div className="capabilities">{manifest.Capabilities?.map((capability) => <span key={capability}>{capability}</span>)}</div>{!installed && <button className="plugin-action" onClick={() => installPlugin(manifest)}>{text.install}</button>}{installation && <div className="lifecycle-actions">{!installation.Ready && <button onClick={() => runPluginAction("plugin.readiness", installation)}>{text.markReady}</button>}{!installation.Enabled && installation.Ready && <button onClick={() => runPluginAction("plugin.enable", installation)}>{text.enable}</button>}{installation.Ready && <small>{text.ready}</small>}</div>}</article>; }) : <p className="muted">{text.empty}</p>}</div>
        </section>
        <section className={`section ${activePage !== "plugins" ? "page-hidden" : ""}`} id="installed-plugin-table"><div className="section-heading"><div><p className="eyebrow">{text.plugins}</p><h2>{text.installed}</h2></div><span className="count">{installations.length}</span></div><div className="table-wrap"><table><thead><tr><th>Plugin</th><th>Version</th><th>Status</th><th>View</th></tr></thead><tbody>{installations.map((installation) => <tr key={`${installation.PluginID}@${installation.Version}`}><td>{installation.PluginID}</td><td>{installation.Version}</td><td>{installation.Enabled && installation.Ready ? text.ready : installation.Ready ? text.disabled : text.notReady}</td><td><a className="view-icon" aria-label={`View ${installation.PluginID}`} title="View plugin" href={`/plugins/${encodeURIComponent(`${installation.PluginID}@${installation.Version}`)}`}>↗</a></td></tr>)}</tbody></table></div></section>
        <section className={`section ${activePage !== "plugins" ? "page-hidden" : ""}`}><div className="section-heading"><div><p className="eyebrow">{text.plugins}</p><h2>{text.registered}</h2></div><span className="count">{manifests.length}</span></div><input className="plugin-search" placeholder={text.search} value={pluginFilter} onChange={(event) => setPluginFilter(event.target.value)} /><div className="plugin-grid">{manifests.filter((manifest) => `${manifest.ID} ${manifest.Title} ${manifest.Description}`.toLowerCase().includes(pluginFilter.toLowerCase())).map((manifest) => { const installed = installations.some((item) => item.PluginID === manifest.ID && item.Version === manifest.Version); return <article className="plugin" key={`catalog-${manifest.ID}@${manifest.Version}`}><div><h3>{manifest.Title || manifest.ID}</h3><p className="muted">{manifest.Description}</p></div><small>{manifest.ID}@{manifest.Version}</small><div className="capabilities">{manifest.Capabilities?.map((capability) => <span key={capability}>{capability}</span>)}</div><button className="plugin-action" disabled={installed} onClick={() => installPlugin(manifest)}>{installed ? text.installedLabel : text.install}</button></article>; })}</div></section>
        <section className={`section ${activePage !== "audit" ? "page-hidden" : ""}`} id="audit"><div className="section-heading"><div><p className="eyebrow">{text.audit}</p><h2>{text.audit}</h2></div><span className="count">{audit.length}</span></div><input className="plugin-search" placeholder={text.searchAudit} value={auditFilter} onChange={(event) => setAuditFilter(event.target.value)} /><div className="audit-list">{audit.length ? audit.slice().reverse().filter((record) => `${record.action} ${record.source} ${record.project_id ?? ""} ${record.plugin_id ?? ""} ${record.detail ?? ""}`.toLowerCase().includes(auditFilter.toLowerCase())).map((record, index) => <div className="audit" key={`${record.action}-${index}`}><span className={record.success ? "success" : "failure"}>{record.success ? "✓" : "!"}</span><div><strong>{record.action}</strong><small>{record.source}{record.plugin_id ? ` · ${record.plugin_id}@${record.version}` : ""}{record.project_id ? ` · ${record.project_id}` : ""}</small></div></div>) : <p className="muted">{text.noAudit}</p>}</div></section>
        <section className={`section ${activePage !== "system" ? "page-hidden" : ""}`} id="system"><div className="section-heading"><div><p className="eyebrow">{text.system}</p><h2>{text.system}</h2></div><span className="count">{system.length}</span></div><div className="plugin-grid">{system.map((manifest) => <article className="plugin" key={`system-${manifest.ID}@${manifest.Version}`}><h3>{manifest.Title || manifest.ID}</h3><p className="muted">{manifest.Description}</p><small>{manifest.Routes?.length ?? 0} {text.routes} · {manifest.ExtensionPoints?.length ?? 0} {text.extensions} · {manifest.SourceFiles?.length ?? 0} {text.files}</small><div className="capabilities">{manifest.Capabilities?.map((capability) => <span key={capability}>{capability}</span>)}</div><div className="route-list">{manifest.Routes?.map((route) => <div key={route.Operation}><code>{route.Operation}</code><small>{route.Capability}</small></div>)}</div><details><summary>{text.files} / {text.imports} / {text.dependencies}</summary><small>{manifest.SourceFiles?.join(" · ")}</small><small>{manifest.Imports?.join(" · ")}</small><small>{[...(manifest.Dependencies ?? []), ...(manifest.PluginImports ?? [])].join(" · ") || "—"}</small></details></article>)}</div></section>
        <section className={`section ${activePage !== "database" ? "page-hidden" : ""}`}><div className="section-heading"><div><p className="eyebrow">{text.database}</p><h2>{text.database}</h2></div></div><div className="plugin-grid">{system.filter((manifest) => manifest.ID === "tables").map((manifest) => <article className="plugin" key={`database-${manifest.ID}`}><h3>{manifest.Title}</h3><p className="muted">{manifest.Description}</p><div className="route-list">{manifest.Routes?.map((route) => <div key={route.Operation}><code>{route.Operation}</code><small>{route.Capability}</small></div>)}</div></article>)}</div></section>
        <section className={`section ${activePage !== "functions" ? "page-hidden" : ""}`}><div className="section-heading"><div><p className="eyebrow">{text.functions}</p><h2>Custom serverless functions</h2><p className="muted">Create project-scoped TypeScript function drafts, test source validation, and save them for host compilation.</p></div></div><div className="function-editor"><label>Function name<input value={functionName} onChange={(event) => setFunctionName(event.target.value)} placeholder="hello" /></label><label>Source code<textarea value={functionSource} onChange={(event) => setFunctionSource(event.target.value)} rows={18} spellCheck={false} /></label><div className="playground-actions"><button type="button" onClick={() => void submitFunction("functions.test")}>Test</button><button type="button" onClick={() => void submitFunction("functions.create")}>Save draft</button><small>{functionMessage}</small></div></div><div className="plugin-grid">{system.filter((manifest) => manifest.ID === "functions").map((manifest) => <article className="plugin" key={`functions-${manifest.ID}`}><h3>{manifest.Title}</h3><p className="muted">{manifest.Description}</p><div className="route-list">{manifest.Routes?.map((route) => <div key={route.Operation}><code>{route.Operation}</code><small>{route.Capability}</small></div>)}</div></article>)}</div></section>
        <section className={`section ${activePage !== "workflows" ? "page-hidden" : ""}`}><div className="section-heading"><div><p className="eyebrow">{text.workflows}</p><h2>{text.workflows}</h2></div></div><div className="plugin-grid">{system.filter((manifest) => manifest.ID === "functions").map((manifest) => <article className="plugin" key={`workflows-${manifest.ID}`}><h3>{manifest.Title}</h3><p className="muted">{manifest.Description}</p><div className="route-list">{manifest.Routes?.filter((route) => route.Operation.includes("workflow") || route.Operation.includes("deploy") || route.Operation.includes("function")).map((route) => <div key={route.Operation}><code>{route.Operation}</code><small>{route.Capability}</small></div>)}</div></article>)}</div></section>
        <section className={`section ${activePage !== "files" ? "page-hidden" : ""}`}><div className="section-heading"><div><p className="eyebrow">{text.files}</p><h2>{text.files}</h2></div></div><div className="plugin-grid">{system.filter((manifest) => manifest.ID === "files").map((manifest) => <article className="plugin" key={`files-${manifest.ID}`}><h3>{manifest.Title}</h3><p className="muted">{manifest.Description}</p><div className="route-list">{manifest.Routes?.map((route) => <div key={route.Operation}><code>{route.Operation}</code><small>{route.Capability}</small></div>)}</div></article>)}</div></section>
        <section className={`section ${activePage !== "live-sessions" ? "page-hidden" : ""}`}><div className="section-heading"><div><p className="eyebrow">{text.liveSessions}</p><h2>{text.liveSessions}</h2></div></div><div className="plugin-grid">{system.filter((manifest) => ["rooms", "matchmaking"].includes(manifest.ID)).map((manifest) => <article className="plugin" key={`live-${manifest.ID}`}><h3>{manifest.Title}</h3><p className="muted">{manifest.Description}</p><div className="route-list">{manifest.Routes?.map((route) => <div key={route.Operation}><code>{route.Operation}</code><small>{route.Capability}</small></div>)}</div></article>)}</div></section>
        <section className={`section ${activePage !== "routes" ? "page-hidden" : ""}`}><div className="section-heading"><div><p className="eyebrow">{text.playground}</p><h2>{text.playground}</h2></div></div><div className="playground"><label>{text.selectRoute}<select value={playgroundRoute} onChange={(event) => setPlaygroundRoute(event.target.value)}><option value="">—</option>{system.flatMap((manifest) => manifest.Routes?.map((route) => <option key={`${manifest.ID}:${route.Operation}`} value={`${manifest.ID}:${route.Operation}`}>{manifest.ID} · {route.Operation}</option>) ?? [])}</select></label><label>{text.payload}<textarea value={playgroundPayload} onChange={(event) => setPlaygroundPayload(event.target.value)} rows={8} /></label><div className="playground-actions"><button type="button" onClick={testPlayground}>{text.test}</button><button type="button" onClick={sendPlayground}>{text.send}</button><small>{playgroundMessage}</small></div><pre>{playgroundRoute ? JSON.stringify({ project_id: selectedProjectID, operation: playgroundRoute.split(":")[1], payload: playgroundPayload }, null, 2) : text.selectRoute}</pre></div><div className="section-heading"><div><p className="eyebrow">{text.routesPage}</p><h2>{text.routesPage}</h2></div><span className="count">{system.reduce((total, manifest) => total + (manifest.Routes?.length ?? 0), 0)}</span></div><div className="plugin-grid">{system.map((manifest) => <article className="plugin" key={`routes-${manifest.ID}`}><h3>{manifest.Title}</h3><div className="route-list">{manifest.Routes?.map((route) => <div key={route.Operation}><code>{route.Operation}</code><small>{route.Capability}</small></div>)}</div></article>)}</div></section>
        <section className={`section ${activePage !== "monitoring" ? "page-hidden" : ""}`}><div className="section-heading"><div><p className="eyebrow">{text.monitoring}</p><h2>{text.monitoring}</h2></div><span className="count">{manifests.length}</span></div><div className="audit-list">{manifests.map((manifest) => { const installation = installations.find((candidate) => candidate.PluginID === manifest.ID && candidate.Version === manifest.Version); const state = !installation ? "not installed" : installation.Draining ? "draining" : !installation.Ready ? text.notReady : !installation.Enabled ? text.disabled : text.ready; const healthy = Boolean(installation?.Enabled && installation?.Ready && !installation.Draining); const statusClass = !installation ? "neutral" : healthy ? "success" : "failure"; return <div className="audit" key={`monitor-${manifest.ID}@${manifest.Version}`}><span className={statusClass}>{healthy ? "✓" : !installation ? "—" : "!"}</span><div><strong>{manifest.Title || manifest.ID}</strong><small>{manifest.ID}@{manifest.Version} · {state}</small></div></div>; })}</div></section>
        <section className="page-hidden"><div className="playground" /></section>
        </section>
      </div>
    </main>
  );
}
