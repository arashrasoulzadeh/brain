"use client";
import { FormEvent, useEffect, useState } from "react";
import { usePathname } from "next/navigation";
import Link from "next/link";
import { AdminFooter, AdminHeader, AdminSidebar } from "./components/admin-chrome";
type Project = { ID: string; Active: boolean };
type PluginManifest = { ID: string; Title: string; Description: string; Version: string; Capabilities: string[]; ConfigSchema?: Record<string, string> };
type PluginStatus = { ProjectID: string; PluginID: string; Version: string; Enabled: boolean; Ready: boolean; WorkerReady?: boolean; Draining: boolean };
type AuditRecord = { source: string; action: string; actor_id?: string; project_id?: string; plugin_id?: string; version?: string; success: boolean; detail?: string };
type FunctionDraft = { project_id: string; name: string; source: string; created_at: string };
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
    workers: "Workers",
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
    workers: "کارگرها",
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
  const [functionSource, setFunctionSource] = useState("export const config = { method: 'POST' };\n\nexport default async function handler(input) {\n  return input;\n}\n");
  const [functionMessage, setFunctionMessage] = useState("");
  const [functionInput, setFunctionInput] = useState("{}");
  const [functionVersion, setFunctionVersion] = useState("1");
  const [functionPublic, setFunctionPublic] = useState(false);
  const [functionPermissions, setFunctionPermissions] = useState("functions.invoke");
  const [functionDrafts, setFunctionDrafts] = useState<FunctionDraft[]>([]);
  const [databaseOperation, setDatabaseOperation] = useState("tables.schema");
  const [databasePayload, setDatabasePayload] = useState("{}");
  const [databaseMessage, setDatabaseMessage] = useState("");
  const [fileOperation, setFileOperation] = useState("files.metadata");
  const [filePayload, setFilePayload] = useState("{\"key\":\"example.txt\"}");
  const [fileMessage, setFileMessage] = useState("");
  const [workflowOperation, setWorkflowOperation] = useState("functions.invoke");
  const [workflowPayload, setWorkflowPayload] = useState("{}");
  const [workflowMessage, setWorkflowMessage] = useState("");
  const [sessionOperation, setSessionOperation] = useState("rooms.create");
  const [sessionPayload, setSessionPayload] = useState("{}");
  const [sessionMessage, setSessionMessage] = useState("");
  const text = copy[language];
  const pageTitle = ({ projects: text.projects, plugins: text.plugins, system: text.system, audit: text.audit, database: text.database, functions: text.functions, workflows: text.workflows, files: text.files, "live-sessions": text.liveSessions, routes: text.routesPage, monitoring: text.monitoring, workers: text.workers } as Record<string, string>)[activePage] ?? text.projects;
  const activePluginIDs = new Set(installations.filter((installation) => installation.Enabled && installation.Ready).map((installation) => installation.PluginID));
  const requiredPlugin = ({ system: "dashboard", database: "tables", functions: "functions", workflows: "functions", files: "files", "live-sessions": "rooms" } as Record<string, string | undefined>)[activePage];
  const pageRequiresAnyPlugin = ["routes", "monitoring", "workers"].includes(activePage);
  useEffect(() => {
    if (!session) return;
    const refresh = () => Promise.all([loadProjects(session), loadPlugins(session, selectedProjectID), loadAudit(session), loadSystem(session), loadFunctionDrafts(session, selectedProjectID)]).catch(() => undefined);
    const timer = window.setInterval(refresh, 10000);
    return () => window.clearInterval(timer);
  }, [session, selectedProjectID]);
  useEffect(() => {
    const savedSession = window.sessionStorage.getItem("brain-admin-session");
    if (!savedSession) { setHydratingSession(false); return; }
    const hydrationTimeout = window.setTimeout(() => {
      window.sessionStorage.removeItem("brain-admin-session");
      setSession("");
      setError("The operator session could not be restored. Please sign in again.");
      setHydratingSession(false);
    }, 5000);
    setSession(savedSession);
    loadProjects(savedSession).then((loadedProjects) => {
      const activeProjectID = loadedProjects.find((project) => project.Active)?.ID ?? "";
      setSelectedProjectID(activeProjectID);
      return Promise.all([loadPlugins(savedSession, activeProjectID), loadAudit(savedSession), loadSystem(savedSession), loadFunctionDrafts(savedSession, activeProjectID)]);
    }).catch(() => {
      window.sessionStorage.removeItem("brain-admin-session");
      setSession("");
      setError("Your operator session expired. Please sign in again.");
    }).finally(() => { window.clearTimeout(hydrationTimeout); setHydratingSession(false); });
    return () => window.clearTimeout(hydrationTimeout);
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
  async function loadFunctionDrafts(token: string, projectID = selectedProjectID) {
    if (!projectID) { setFunctionDrafts([]); return; }
    const response = await fetch(`/api/admin/functions?project_id=${encodeURIComponent(projectID)}`, { headers: { authorization: `Bearer ${token}` } });
    if (response.ok) setFunctionDrafts(await response.json() as FunctionDraft[]);
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
      if (!response.ok) {
        const body = await response.text();
        let message = body || text.invalid;
        try {
          const parsed = JSON.parse(body) as { message?: string };
          message = parsed.message || text.invalid;
        } catch {
          // The host should return JSON, but preserve useful legacy text during upgrades.
        }
        throw new Error(message);
      }
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
  async function sendPlayground() {
    testPlayground();
    if (!playgroundRoute) { setPlaygroundMessage("Select a route first"); return; }
    if (!selectedProjectID) { setPlaygroundMessage("Select an active project first"); return; }
    const [pluginID, targetOperation] = playgroundRoute.split(":");
    const manifest = system.find((candidate) => candidate.ID === pluginID);
    if (!manifest) { setPlaygroundMessage("Plugin manifest not found"); return; }
    let payload: unknown;
    try { payload = JSON.parse(playgroundPayload || "{}"); } catch { setPlaygroundMessage("Invalid JSON payload"); return; }
    setPlaygroundMessage("Executing through the host...");
    const response = await fetch("/api/admin/actions", { method: "POST", headers: { authorization: `Bearer ${session}`, "content-type": "application/json" }, body: JSON.stringify({ operation: "plugin.invoke", project_id: selectedProjectID, plugin_id: pluginID, version: manifest.Version, target_operation: targetOperation, payload }) });
    if (!response.ok) { setPlaygroundMessage(await response.text()); return; }
    const result = await response.json();
    setPlaygroundMessage(result.success === false ? `Plugin error: ${result.errorCode ?? "execution failed"}` : "Host execution completed");
    let output: unknown = result;
    if (typeof result.payload === "string") {
      try { output = JSON.parse(atob(result.payload)); } catch { output = result.payload; }
    }
    setPlaygroundPayload(JSON.stringify(output, null, 2));
  }
  async function runDatabaseOperation() {
    if (!selectedProjectID) { setDatabaseMessage("Select an active project first"); return; }
    let payload: unknown;
    try { payload = JSON.parse(databasePayload || "{}"); } catch { setDatabaseMessage("Invalid JSON payload"); return; }
    const manifest = system.find((candidate) => candidate.ID === "tables");
    if (!manifest) { setDatabaseMessage("Tables plugin manifest unavailable"); return; }
    setDatabaseMessage("Executing through the host...");
    const response = await fetch("/api/admin/actions", { method: "POST", headers: { authorization: `Bearer ${session}`, "content-type": "application/json" }, body: JSON.stringify({ operation: "plugin.invoke", project_id: selectedProjectID, plugin_id: "tables", version: manifest.Version, target_operation: databaseOperation, payload }) });
    if (!response.ok) { setDatabaseMessage(await response.text()); return; }
    const result = await response.json();
    if (result.success === false) { setDatabaseMessage(`Database plugin error: ${result.errorCode ?? "execution failed"}`); return; }
    try { setDatabasePayload(JSON.stringify(typeof result.payload === "string" ? JSON.parse(atob(result.payload)) : result, null, 2)); } catch { setDatabasePayload(JSON.stringify(result, null, 2)); }
    setDatabaseMessage("Database operation completed");
  }
  async function runFileOperation() {
    if (!selectedProjectID) { setFileMessage("Select an active project first"); return; }
    let payload: unknown;
    try { payload = JSON.parse(filePayload || "{}"); } catch { setFileMessage("Invalid JSON payload"); return; }
    const manifest = system.find((candidate) => candidate.ID === "files");
    if (!manifest) { setFileMessage("Files plugin manifest unavailable"); return; }
    setFileMessage("Executing through the host...");
    const response = await fetch("/api/admin/actions", { method: "POST", headers: { authorization: `Bearer ${session}`, "content-type": "application/json" }, body: JSON.stringify({ operation: "plugin.invoke", project_id: selectedProjectID, plugin_id: "files", version: manifest.Version, target_operation: fileOperation, payload }) });
    if (!response.ok) { setFileMessage(await response.text()); return; }
    const result = await response.json();
    if (result.success === false) { setFileMessage(`Files plugin error: ${result.errorCode ?? "execution failed"}`); return; }
    try { setFilePayload(JSON.stringify(typeof result.payload === "string" ? JSON.parse(atob(result.payload)) : result, null, 2)); } catch { setFilePayload(JSON.stringify(result, null, 2)); }
    setFileMessage("File operation completed");
  }
  async function runWorkflowOperation() {
    await runPluginConsoleOperation("functions", workflowOperation, workflowPayload, setWorkflowPayload, setWorkflowMessage);
  }
  async function runSessionOperation() {
    const pluginID = sessionOperation.startsWith("rooms.") ? "rooms" : "matchmaking";
    await runPluginConsoleOperation(pluginID, sessionOperation, sessionPayload, setSessionPayload, setSessionMessage);
  }
  async function runPluginConsoleOperation(pluginID: string, operation: string, rawPayload: string, setPayload: (value: string) => void, setMessage: (value: string) => void) {
    if (!selectedProjectID) { setMessage("Select an active project first"); return; }
    let payload: unknown;
    try { payload = JSON.parse(rawPayload || "{}"); } catch { setMessage("Invalid JSON payload"); return; }
    const manifest = system.find((candidate) => candidate.ID === pluginID);
    if (!manifest) { setMessage(`${pluginID} plugin manifest unavailable`); return; }
    setMessage("Executing through the host...");
    const response = await fetch("/api/admin/actions", { method: "POST", headers: { authorization: `Bearer ${session}`, "content-type": "application/json" }, body: JSON.stringify({ operation: "plugin.invoke", project_id: selectedProjectID, plugin_id: pluginID, version: manifest.Version, target_operation: operation, payload }) });
    if (!response.ok) { setMessage(await response.text()); return; }
    const result = await response.json();
    if (result.success === false) { setMessage(`Plugin error: ${result.errorCode ?? "execution failed"}`); return; }
    try { setPayload(JSON.stringify(typeof result.payload === "string" ? JSON.parse(atob(result.payload)) : result, null, 2)); } catch { setPayload(JSON.stringify(result, null, 2)); }
    setMessage("Operation completed");
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
  async function submitFunction(operation: "functions.test" | "functions.create" | "functions.run") {
    if (!selectedProjectID || !functionName.trim() || !functionSource.trim()) { setFunctionMessage("Select an active project, name the function, and provide source."); return; }
    let input: unknown = {};
    try { input = JSON.parse(functionInput || "{}"); } catch { setFunctionMessage("Input must be valid JSON."); return; }
    const response = await fetch("/api/admin/functions", { method: "POST", headers: { authorization: `Bearer ${session}`, "content-type": "application/json" }, body: JSON.stringify({ operation, project_id: selectedProjectID, name: functionName, source: functionSource, input, function_version: Number(functionVersion) || 1, public: functionPublic, permissions: functionPermissions.split(",").map((permission) => permission.trim()).filter(Boolean) }) });
    if (!response.ok) { setFunctionMessage(await response.text()); return; }
    const result = await response.json(); setFunctionMessage(operation === "functions.create" ? "Function draft saved." : JSON.stringify(result));
    if (operation === "functions.create") await loadFunctionDrafts(session, selectedProjectID);
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
  if ((requiredPlugin && !activePluginIDs.has(requiredPlugin)) || (pageRequiresAnyPlugin && activePluginIDs.size === 0))
    return <main className="shell" dir={language === "fa" ? "rtl" : "ltr"}><section className="card"><AdminHeader title={pageTitle} eyebrow={text.eyebrow} /><p className="error">This page is unavailable until its plugin is ready and enabled.</p><Link href="/plugins">Open Plugins to enable it</Link><AdminFooter /></section></main>;
  return (
    <main className="shell" dir={language === "fa" ? "rtl" : "ltr"}>
      <div className="workspace">
        <AdminSidebar activePage={activePage} projectID={selectedProjectID} onSignOut={() => void signOut()} visiblePages={["projects", ...(manifests.length ? ["plugins"] : []), "audit", ...(activePluginIDs.has("dashboard") ? ["system"] : []), ...(activePluginIDs.has("tables") ? ["database"] : []), ...(activePluginIDs.has("functions") ? ["functions", "workflows"] : []), ...(activePluginIDs.has("files") ? ["files"] : []), ...(activePluginIDs.has("rooms") || activePluginIDs.has("matchmaking") ? ["live-sessions"] : []), ...(activePluginIDs.size ? ["routes", "monitoring", "workers"] : [])]} />
        <section className="card">
        <AdminHeader title={pageTitle} eyebrow={text.eyebrow} />
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
        <section className={`section ${activePage !== "database" ? "page-hidden" : ""}`}><div className="section-heading"><div><p className="eyebrow">{text.database}</p><h2>{text.database}</h2><p className="muted">Run project-scoped table operations through the installed Tables plugin.</p></div></div><div className="playground"><label>Operation<select value={databaseOperation} onChange={(event) => setDatabaseOperation(event.target.value)}><option value="tables.schema">tables.schema</option><option value="tables.read">tables.read</option><option value="tables.write">tables.write</option></select></label><label>{text.payload}<textarea value={databasePayload} onChange={(event) => setDatabasePayload(event.target.value)} rows={10} /></label><div className="playground-actions"><button type="button" onClick={() => void runDatabaseOperation()}>{text.send}</button><small>{databaseMessage}</small></div></div><div className="plugin-grid">{system.filter((manifest) => manifest.ID === "tables").map((manifest) => <article className="plugin" key={`database-${manifest.ID}`}><h3>{manifest.Title}</h3><p className="muted">{manifest.Description}</p><div className="route-list">{manifest.Routes?.map((route) => <div key={route.Operation}><code>{route.Operation}</code><small>{route.Capability}</small></div>)}</div></article>)}</div></section>
        <section className={`section ${activePage !== "functions" ? "page-hidden" : ""}`}><div className="section-heading"><div><p className="eyebrow">{text.functions}</p><h2>Custom serverless functions</h2><p className="muted">Create project-scoped TypeScript function drafts, test source validation, run JSON input, and save them for host compilation.</p></div></div><div className="function-editor"><label>Function name<input value={functionName} onChange={(event) => setFunctionName(event.target.value)} placeholder="hello" /></label><label>Source code<textarea value={functionSource} onChange={(event) => setFunctionSource(event.target.value)} rows={18} spellCheck={false} /></label><label>JSON input<textarea value={functionInput} onChange={(event) => setFunctionInput(event.target.value)} rows={5} spellCheck={false} /></label><div className="playground-actions"><button type="button" onClick={() => void submitFunction("functions.test")}>Validate</button><button type="button" onClick={() => void submitFunction("functions.run")}>Run test</button><button type="button" onClick={() => void submitFunction("functions.create")}>Save draft</button><small>{functionMessage}</small></div></div><div className="table-wrap"><table><thead><tr><th>Name</th><th>Project</th><th>Created</th><th>Load</th></tr></thead><tbody>{functionDrafts.map((draft) => <tr key={`${draft.project_id}:${draft.name}`}><td><code>{draft.name}</code></td><td>{draft.project_id}</td><td>{new Date(draft.created_at).toLocaleString()}</td><td><button type="button" onClick={() => { setFunctionName(draft.name); setFunctionSource(draft.source); }}>Load</button></td></tr>)}</tbody></table></div><div className="plugin-grid">{system.filter((manifest) => manifest.ID === "functions").map((manifest) => <article className="plugin" key={`functions-${manifest.ID}`}><h3>{manifest.Title}</h3><p className="muted">{manifest.Description}</p><div className="route-list">{manifest.Routes?.map((route) => <div key={route.Operation}><code>{route.Operation}</code><small>{route.Capability}</small></div>)}</div></article>)}</div></section>
        <section className={`section ${activePage !== "workflows" ? "page-hidden" : ""}`}><div className="section-heading"><div><p className="eyebrow">{text.workflows}</p><h2>{text.workflows}</h2><p className="muted">Run project-scoped Functions plugin operations through the host.</p></div></div><div className="playground"><label>Operation<select value={workflowOperation} onChange={(event) => setWorkflowOperation(event.target.value)}><option value="functions.invoke">functions.invoke</option><option value="functions.publish">functions.publish</option></select></label><label>{text.payload}<textarea value={workflowPayload} onChange={(event) => setWorkflowPayload(event.target.value)} rows={10} /></label><div className="playground-actions"><button type="button" onClick={() => void runWorkflowOperation()}>{text.send}</button><small>{workflowMessage}</small></div></div></section>
        <section className={`section ${activePage !== "files" ? "page-hidden" : ""}`}><div className="section-heading"><div><p className="eyebrow">{text.files}</p><h2>{text.files}</h2><p className="muted">Run project-scoped file operations through the installed Files plugin.</p></div></div><div className="playground"><label>Operation<select value={fileOperation} onChange={(event) => setFileOperation(event.target.value)}><option value="files.upload">files.upload</option><option value="files.download">files.download</option><option value="files.metadata">files.metadata</option><option value="files.publish">files.publish</option></select></label><label>{text.payload}<textarea value={filePayload} onChange={(event) => setFilePayload(event.target.value)} rows={10} /></label><div className="playground-actions"><button type="button" onClick={() => void runFileOperation()}>{text.send}</button><small>{fileMessage}</small></div></div><div className="plugin-grid">{system.filter((manifest) => manifest.ID === "files").map((manifest) => <article className="plugin" key={`files-${manifest.ID}`}><h3>{manifest.Title}</h3><p className="muted">{manifest.Description}</p><div className="route-list">{manifest.Routes?.map((route) => <div key={route.Operation}><code>{route.Operation}</code><small>{route.Capability}</small></div>)}</div></article>)}</div></section>
        <section className={`section ${activePage !== "live-sessions" ? "page-hidden" : ""}`}><div className="section-heading"><div><p className="eyebrow">{text.liveSessions}</p><h2>{text.liveSessions}</h2><p className="muted">Manage project-scoped Rooms and Matchmaking operations through the host.</p></div></div><div className="playground"><label>Operation<select value={sessionOperation} onChange={(event) => setSessionOperation(event.target.value)}><option value="rooms.create">rooms.create</option><option value="rooms.join">rooms.join</option><option value="rooms.reconnect">rooms.reconnect</option><option value="rooms.result">rooms.result</option><option value="matchmaking.queue">matchmaking.queue</option><option value="matchmaking.cancel">matchmaking.cancel</option><option value="matchmaking.status">matchmaking.status</option><option value="matchmaking.assign">matchmaking.assign</option></select></label><label>{text.payload}<textarea value={sessionPayload} onChange={(event) => setSessionPayload(event.target.value)} rows={10} /></label><div className="playground-actions"><button type="button" onClick={() => void runSessionOperation()}>{text.send}</button><small>{sessionMessage}</small></div></div><div className="plugin-grid">{system.filter((manifest) => ["rooms", "matchmaking"].includes(manifest.ID)).map((manifest) => <article className="plugin" key={`live-${manifest.ID}`}><h3>{manifest.Title}</h3><p className="muted">{manifest.Description}</p><div className="route-list">{manifest.Routes?.map((route) => <div key={route.Operation}><code>{route.Operation}</code><small>{route.Capability}</small></div>)}</div></article>)}</div></section>
        <section className={`section ${activePage !== "routes" ? "page-hidden" : ""}`}><div className="section-heading"><div><p className="eyebrow">{text.playground}</p><h2>{text.playground}</h2></div></div><div className="playground"><label>{text.selectRoute}<select value={playgroundRoute} onChange={(event) => setPlaygroundRoute(event.target.value)}><option value="">—</option>{system.flatMap((manifest) => manifest.Routes?.map((route) => <option key={`${manifest.ID}:${route.Operation}`} value={`${manifest.ID}:${route.Operation}`}>{manifest.ID} · {route.Operation}</option>) ?? [])}</select></label><label>{text.payload}<textarea value={playgroundPayload} onChange={(event) => setPlaygroundPayload(event.target.value)} rows={8} /></label><div className="playground-actions"><button type="button" onClick={testPlayground}>{text.test}</button><button type="button" onClick={sendPlayground}>{text.send}</button><small>{playgroundMessage}</small></div><pre>{playgroundRoute ? JSON.stringify({ project_id: selectedProjectID, operation: playgroundRoute.split(":")[1], payload: playgroundPayload }, null, 2) : text.selectRoute}</pre></div><div className="section-heading"><div><p className="eyebrow">{text.routesPage}</p><h2>{text.routesPage}</h2></div><span className="count">{system.reduce((total, manifest) => total + (manifest.Routes?.length ?? 0), 0)}</span></div><div className="plugin-grid">{system.map((manifest) => <article className="plugin" key={`routes-${manifest.ID}`}><h3>{manifest.Title}</h3><div className="route-list">{manifest.Routes?.map((route) => <div key={route.Operation}><code>{route.Operation}</code><small>{route.Capability}</small></div>)}</div></article>)}</div></section>
        <section className={`section ${activePage !== "monitoring" ? "page-hidden" : ""}`}><div className="section-heading"><div><p className="eyebrow">{text.monitoring}</p><h2>{text.monitoring}</h2></div><span className="count">{manifests.length}</span></div><div className="audit-list">{manifests.map((manifest) => { const installation = installations.find((candidate) => candidate.PluginID === manifest.ID && candidate.Version === manifest.Version); const state = !installation ? "not installed" : installation.Draining ? "draining" : !installation.Ready ? text.notReady : !installation.Enabled ? text.disabled : installation.WorkerReady === false ? "worker unavailable" : text.ready; const healthy = Boolean(installation?.Enabled && installation?.Ready && installation.WorkerReady !== false && !installation.Draining); const statusClass = !installation ? "neutral" : healthy ? "success" : "failure"; return <div className="audit" key={`monitor-${manifest.ID}@${manifest.Version}`}><span className={statusClass}>{healthy ? "✓" : !installation ? "—" : "!"}</span><div><strong>{manifest.Title || manifest.ID}</strong><small>{manifest.ID}@{manifest.Version} · {state}</small></div></div>; })}</div></section>
        <section className="page-hidden"><div className="playground" /></section>
        <AdminFooter />
        </section>
      </div>
    </main>
  );
}
