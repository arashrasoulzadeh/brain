"use client";
import { adminHeaders, readAdminSession, saveAdminSession, clearAdminSession, signOutAdminSession, saveAdminProject, preferredAdminProject, projectAdminURL, systemAdminURL, projectUIURL, systemUIURL, adminUIContext } from "./lib/admin-api";
import { FormEvent, useEffect, useState } from "react";
import { usePathname } from "next/navigation";
import Link from "next/link";
import { AdminFooter, AdminHeader, AdminSidebar } from "./components/admin-chrome";
import { QUIET_REQUEST_HEADER } from "./components/toast-provider";
import type { Project, PluginManifest, PluginStatus, AuditRecord, FunctionDraft, SystemManifest, Language, ProjectHTTPRoute } from "./lib/admin-types";
import { copy } from "./lib/admin-copy";
import { createProjectRequestGate } from "./lib/project-requests";
import { canViewAdminPage, enabledPluginIDs, visibleAdminPages } from "./lib/admin-navigation";
import { AuditPanel } from "./components/audit-panel";
import { SystemPanel } from "./components/system-panel";
import { MonitoringPanel } from "./components/monitoring-panel";
import { ProjectsPanel } from "./components/projects-panel";
import { PluginCatalogPanel } from "./components/plugin-catalog-panel";
import { FunctionDraftsPanel } from "./components/function-drafts-panel";
import { RoutesPanel } from "./components/routes-panel";
export default function AdminPage() {
  const pathname = usePathname();
  const { page: activePage, namespace } = adminUIContext(pathname);
  const [language, setLanguage] = useState<Language>("en");
  const [session, setSession] = useState("");
  const [hydratingSession, setHydratingSession] = useState(true);
  const [projects, setProjects] = useState<Project[]>([]);
  const [manifests, setManifests] = useState<PluginManifest[]>([]);
  const [installations, setInstallations] = useState<PluginStatus[]>([]);
  const [projectRoutes, setProjectRoutes] = useState<ProjectHTTPRoute[]>([]);
  const [audit, setAudit] = useState<AuditRecord[]>([]);
  const [system, setSystem] = useState<SystemManifest[]>([]);
  const [username, setUsername] = useState("");
  const [secret, setSecret] = useState("");
  const [projectID, setProjectID] = useState("");
  const [projectName, setProjectName] = useState("");
  const [projectSlug, setProjectSlug] = useState("");
  const [selectedProjectID, setSelectedProjectID] = useState("");
  const [projectRequests] = useState(createProjectRequestGate);
  const [error, setError] = useState("");
  const [configuration, setConfiguration] = useState<
    Record<string, Record<string, string | boolean>>
  >({});
  const [pluginFilter, setPluginFilter] = useState("");
  const [auditFilter, setAuditFilter] = useState("");
  const [playgroundRoute, setPlaygroundRoute] = useState("");
  const [playgroundPayload, setPlaygroundPayload] = useState("{}");
  const [playgroundMessage, setPlaygroundMessage] = useState("");
  const [functionName, setFunctionName] = useState("");
  const [functionSource, setFunctionSource] = useState(
    "export const config = { method: 'POST' };\n\nexport default async function handler(input) {\n  return input;\n}\n",
  );
  const [functionMessage, setFunctionMessage] = useState("");
  const [functionInput, setFunctionInput] = useState("{}");
  const [functionVersion, setFunctionVersion] = useState("1");
  const [functionPublic, setFunctionPublic] = useState(false);
  const [functionPermissions, setFunctionPermissions] = useState("functions.invoke");
  const [functionDrafts, setFunctionDrafts] = useState<FunctionDraft[]>([]);
  const text = copy[language];
  const pageTitle =
    (
      {
        projects: text.projects,
        plugins: text.plugins,
        system: text.system,
        audit: text.audit,
        functions: text.functions,
        routes: text.routesPage,
        monitoring: text.monitoring,
        workers: text.workers,
      } as Record<string, string>
    )[activePage] ?? text.projects;
  const activePluginIDs = enabledPluginIDs(installations);
  function selectProject(id: string) {
    if (!projectRequests.select(id)) return;
    setSelectedProjectID(id);
    setInstallations([]);
    setProjectRoutes([]);
    setFunctionDrafts([]);
    setPlaygroundRoute("");
    setPlaygroundPayload("{}");
    setPlaygroundMessage("");
    setFunctionMessage("");
    saveAdminProject(id);
  }
  useEffect(() => {
    if (!session) return;
    const refresh = () =>
      Promise.all([
        loadProjects(session),
        loadPlugins(session, selectedProjectID),
        loadAudit(session),
        loadSystem(session),
        loadFunctionDrafts(session, selectedProjectID),
      ]).catch(() => undefined);
    const timer = window.setInterval(refresh, 10000);
    return () => window.clearInterval(timer);
  }, [session, selectedProjectID]);
  useEffect(() => {
    const savedSession = readAdminSession();
    if (!savedSession) {
      setHydratingSession(false);
      return;
    }
    const hydrationTimeout = window.setTimeout(() => {
      clearAdminSession();
      setSession("");
      setError("The operator session could not be restored. Please sign in again.");
      setHydratingSession(false);
    }, 5000);
    setSession(savedSession);
    loadProjects(savedSession)
      .then((loadedProjects) => {
        const activeProjectID = preferredAdminProject(loadedProjects, namespace)?.ID ?? "";
        selectProject(activeProjectID);
        return Promise.all([
          loadPlugins(savedSession, activeProjectID),
          loadAudit(savedSession),
          loadSystem(savedSession),
          loadFunctionDrafts(savedSession, activeProjectID),
        ]);
      })
      .catch(() => {
        clearAdminSession();
        setSession("");
        setError("Your operator session expired. Please sign in again.");
      })
      .finally(() => {
        window.clearTimeout(hydrationTimeout);
        setHydratingSession(false);
      });
    return () => window.clearTimeout(hydrationTimeout);
  }, [namespace]);
  async function loadProjects(token: string): Promise<Project[]> {
    const response = await fetch(systemAdminURL("projects"), {
      headers: adminHeaders(token),
    });
    if (!response.ok) throw new Error("Could not load projects");
    const loadedProjects = (await response.json()) as Project[];
    setProjects(loadedProjects);
    return loadedProjects;
  }
  async function loadProjectRoutes(token: string, project: string) {
    if (!project) {
      setProjectRoutes([]);
      return;
    }
    const response = await fetch(projectAdminURL(project, "routes"), { headers: adminHeaders(token) });
    if (!response.ok) throw new Error("Could not discover project routes");
    const routes = (await response.json()) as ProjectHTTPRoute[] | null;
    if (projectRequests.current() === project && readAdminSession() === token) {
      setProjectRoutes(routes ?? []);
    }
  }
  async function loadPlugins(token: string, projectID = selectedProjectID) {
    if (!projectID) { setManifests([]); setInstallations([]); setProjectRoutes([]); return; }
    const current = projectRequests.begin(projectID, "catalog");
    if (!current) return;
    const valid = () => current() && readAdminSession() === token;
    try {
      const response = await fetch(projectAdminURL(projectID, "plugins"), { headers: adminHeaders(token) });
      if (!valid()) return;
      if (!response.ok) throw new Error("Could not load plugins");
      const result = (await response.json()) as { manifests: PluginManifest[] | null; installations: PluginStatus[] | null };
      if (!valid()) return;
      setManifests(result.manifests ?? []);
      setInstallations((result.installations ?? []).filter(item => item.ProjectID === projectID));
      await loadProjectRoutes(token, projectID).catch(() => setProjectRoutes([]));
    } catch (reason) {
      if (!valid()) return;
      setInstallations([]);
      throw reason;
    }
  }
  async function loadAudit(token: string) {
    const response = await fetch(systemAdminURL("audit"), {
      headers: adminHeaders(token),
    });
    if (!response.ok) throw new Error("Could not load audit log");
    setAudit((await response.json()) as AuditRecord[]);
  }
  async function loadSystem(token: string) {
    const response = await fetch(systemAdminURL("system"), {
      headers: adminHeaders(token),
    });
    if (!response.ok) throw new Error("Could not load system capabilities");
    setSystem((await response.json()) as SystemManifest[]);
  }
  async function loadFunctionDrafts(token: string, projectID = selectedProjectID) {
    const current = projectRequests.begin(projectID, "functions");
    if (!current) return;
    const valid = () => current() && readAdminSession() === token;
    if (!projectID) { setFunctionDrafts([]); return; }
    try {
      const response = await fetch(
        projectAdminURL(projectID, "functions"),
        { headers: adminHeaders(token, { [QUIET_REQUEST_HEADER]: "1" }) },
      );
      const result = response.ok ? (await response.json()) as FunctionDraft[] | null : [];
      if (valid()) setFunctionDrafts((result ?? []).filter(item => item.project_id === projectID));
    } catch {
      if (valid()) setFunctionDrafts([]);
    }
  }
  async function signIn(event: FormEvent) {
    event.preventDefault();
    setError("");
    try {
      const response = await fetch(systemAdminURL("session"), {
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
      saveAdminSession(token);
      const loadedProjects = await loadProjects(token);
      const activeProject = preferredAdminProject(loadedProjects, namespace)?.ID ?? "";
      selectProject(activeProject);
      await Promise.all([loadPlugins(token, activeProject), loadFunctionDrafts(token, activeProject)]);
      await loadAudit(token);
      await loadSystem(token);
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Sign in failed");
    }
  }
  async function signOut() {
    selectProject("");
    await signOutAdminSession();
    setSession("");
  }
  async function createProject(event: FormEvent) {
    event.preventDefault();
    setError("");
    try {
      const response = await fetch(systemAdminURL("actions"), {
        method: "POST",
        headers: adminHeaders(session, { "content-type": "application/json" }),
        body: JSON.stringify({ operation: "project.create", project_id: projectID, name: projectName, slug: projectSlug }),
      });
      if (!response.ok) throw new Error(await response.text());
      setProjectID("");
      setProjectName("");
      setProjectSlug("");
      const loaded = await loadProjects(session);
      if (!projectRequests.current()) {
        const next = loaded.find(candidate => candidate.ID === projectID && candidate.Active)?.ID ?? loaded.find(candidate => candidate.Active)?.ID ?? "";
        selectProject(next);
        await Promise.all([loadPlugins(session, next), loadFunctionDrafts(session, next)]);
      }
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Could not create project");
    }
  }
  async function configureProject(project: Project) {
    const name = window.prompt("Display name", project.Name || project.ID);
    if (name === null) return;
    const slug = window.prompt("Unique URL slug", project.Slug || project.ID);
    if (slug === null) return;
    const rawSettings = window.prompt("Project settings (JSON object)", JSON.stringify(project.Settings ?? {}, null, 2));
    if (rawSettings === null) return;
    try {
      const values: unknown = JSON.parse(rawSettings);
      if (!values || typeof values !== "object" || Array.isArray(values)) throw new Error("Settings must be a JSON object");
      const response = await fetch(systemAdminURL("actions"), {
        method: "POST",
        headers: adminHeaders(session, { "content-type": "application/json" }),
        body: JSON.stringify({ operation: "project.configure", project_id: project.ID, name, slug, values }),
      });
      if (!response.ok) throw new Error(await response.text());
      await loadProjects(session);
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Could not update project");
    }
  }
  async function setProjectActive(project: Project, active: boolean) {
    if (!active && !window.confirm(`${text.deactivate} ${project.ID}?`)) return;
    const response = await fetch(systemAdminURL("actions"), {
      method: "POST",
      headers: adminHeaders(session, { "content-type": "application/json" }),
      body: JSON.stringify({
        operation: active ? "project.activate" : "project.deactivate",
        project_id: project.ID,
        confirmed: !active,
      }),
    });
    if (!response.ok) {
      setError(await response.text());
      return;
    }
    const loadedProjects = await loadProjects(session);
    const current = projectRequests.current();
    const next = !current || (current === project.ID && !active)
      ? loadedProjects.find(candidate => candidate.Active)?.ID ?? ""
      : current;
    selectProject(next);
    await Promise.all([loadPlugins(session, next), loadFunctionDrafts(session, next)]);
    await loadAudit(session);
  }
  function testPlayground() {
    try {
      JSON.parse(playgroundPayload || "{}");
      setPlaygroundMessage("Valid JSON payload");
    } catch {
      setPlaygroundMessage("Invalid JSON payload");
    }
  }
  async function sendPlayground() {
    testPlayground();
    if (!playgroundRoute) {
      setPlaygroundMessage("Select a route first");
      return;
    }
    if (!selectedProjectID) {
      setPlaygroundMessage("Select an active project first");
      return;
    }
    const [pluginID, targetOperation] = playgroundRoute.split(":");
    const manifest = system.find((candidate) => candidate.ID === pluginID);
    if (!manifest) {
      setPlaygroundMessage("Plugin manifest not found");
      return;
    }
    let payload: unknown;
    try {
      payload = JSON.parse(playgroundPayload || "{}");
    } catch {
      setPlaygroundMessage("Invalid JSON payload");
      return;
    }
    setPlaygroundMessage("Executing through the host...");
    const response = await fetch(projectAdminURL(selectedProjectID, "actions"), {
      method: "POST",
      headers: adminHeaders(session, { "content-type": "application/json" }),
      body: JSON.stringify({
        operation: "plugin.invoke",
        project_id: selectedProjectID,
        plugin_id: pluginID,
        version: manifest.Version,
        target_operation: targetOperation,
        payload,
      }),
    });
    if (!response.ok) {
      setPlaygroundMessage(await response.text());
      return;
    }
    const result = await response.json();
    setPlaygroundMessage(
      result.success === false
        ? `Plugin error: ${result.errorCode ?? "execution failed"}`
        : "Host execution completed",
    );
    let output: unknown = result;
    if (typeof result.payload === "string") {
      try {
        output = JSON.parse(atob(result.payload));
      } catch {
        output = result.payload;
      }
    }
    setPlaygroundPayload(JSON.stringify(output, null, 2));
  }
  async function installPlugin(manifest: PluginManifest) {
    setError("");
    const project = projects.find(
      (candidate) => candidate.ID === selectedProjectID && candidate.Active,
    );
    if (!project) {
      setError("Create an active project first");
      return;
    }
    try {
      const response = await fetch(projectAdminURL(project.ID, "actions"), {
        method: "POST",
        headers: adminHeaders(session, { "content-type": "application/json" }),
        body: JSON.stringify({
          operation: "plugin.install",
          project_id: project.ID,
          plugin_id: manifest.ID,
          version: manifest.Version,
        }),
      });
      if (!response.ok) throw new Error(await response.text());
      await loadPlugins(session, selectedProjectID);
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Could not install plugin");
    }
  }
  async function submitFunction(
    operation: "functions.test" | "functions.create" | "functions.run",
  ) {
    if (!selectedProjectID || !functionName.trim() || !functionSource.trim()) {
      setFunctionMessage("Select an active project, name the function, and provide source.");
      return;
    }
    let input: unknown = {};
    try {
      input = JSON.parse(functionInput || "{}");
    } catch {
      setFunctionMessage("Input must be valid JSON.");
      return;
    }
    const response = await fetch(projectAdminURL(selectedProjectID, "functions"), {
      method: "POST",
      headers: adminHeaders(session, { "content-type": "application/json" }),
      body: JSON.stringify({
        operation,
        project_id: selectedProjectID,
        name: functionName,
        source: functionSource,
        input,
        function_version: Number(functionVersion) || 1,
        public: functionPublic,
        permissions: functionPermissions
          .split(",")
          .map((permission) => permission.trim())
          .filter(Boolean),
      }),
    });
    if (!response.ok) {
      setFunctionMessage(await response.text());
      return;
    }
    const result = await response.json();
    setFunctionMessage(
      operation === "functions.create" ? "Function draft saved." : JSON.stringify(result),
    );
    if (operation === "functions.create") await loadFunctionDrafts(session, selectedProjectID);
  }
  async function uninstallProject(project: Project) {
    if (!window.confirm(`${text.uninstall} ${project.ID}?`)) return;
    setError("");
    const response = await fetch(systemAdminURL("actions"), {
      method: "POST",
      headers: adminHeaders(session, { "content-type": "application/json" }),
      body: JSON.stringify({
        operation: "project.uninstall",
        project_id: project.ID,
        confirmed: true,
      }),
    });
    if (!response.ok) {
      setError(await response.text());
      return;
    }
    const loadedProjects = await loadProjects(session);
    const current = projectRequests.current();
    const next = loadedProjects.some(candidate => candidate.ID === current && candidate.Active)
      ? current : loadedProjects.find(candidate => candidate.Active)?.ID ?? "";
    selectProject(next);
    await Promise.all([loadPlugins(session, next), loadFunctionDrafts(session, next)]);
    await loadAudit(session);
  }
  async function runPluginAction(
    operation: string,
    installation: PluginStatus,
    requestedCapability?: string,
  ) {
    setError("");
    const project = projects.find(
      (candidate) => candidate.ID === selectedProjectID && candidate.Active,
    );
    if (!project) {
      setError("Create an active project first");
      return;
    }
    const manifest = manifests.find(
      (candidate) =>
        candidate.ID === installation.PluginID && candidate.Version === installation.Version,
    );
    const capability = requestedCapability ?? manifest?.Capabilities?.[0] ?? "";
    if (
      (operation === "plugin.remove" ||
        operation === "plugin.revoke" ||
        operation === "plugin.rollback" ||
        operation === "plugin.stop" ||
        operation === "plugin.stop-immediate") &&
      !window.confirm(`${operation} ${installation.PluginID}@${installation.Version}?`)
    )
      return;
    const response = await fetch(operation === "plugin.readiness" ? systemAdminURL("actions") : projectAdminURL(project.ID, "actions"), {
      method: "POST",
      headers: adminHeaders(session, { "content-type": "application/json" }),
      body: JSON.stringify({
        operation,
        project_id: project.ID,
        plugin_id: installation.PluginID,
        version: installation.Version,
        capability,
        ready: operation === "plugin.readiness",
        confirmed: [
          "plugin.remove",
          "plugin.revoke",
          "plugin.rollback",
          "plugin.stop",
          "plugin.stop-immediate",
        ].includes(operation),
      }),
    });
    if (!response.ok) {
      setError(await response.text());
      return;
    }
    await loadPlugins(session, selectedProjectID);
    await loadAudit(session);
  }
  async function configurePlugin(installation: PluginStatus) {
    if (!selectedProjectID) {
      setError("Select an active project first");
      return;
    }
    const key = `${installation.PluginID}@${installation.Version}`;
    const manifest = manifests.find(
      (candidate) =>
        candidate.ID === installation.PluginID && candidate.Version === installation.Version,
    );
    const rawValues = configuration[key] ?? {};
    const values = Object.fromEntries(
      Object.entries(rawValues).map(([name, value]) => [
        name,
        manifest?.ConfigSchema?.[name] === "int" ? Number(value) : value,
      ]),
    );
    const response = await fetch(projectAdminURL(selectedProjectID, "actions"), {
      method: "POST",
      headers: adminHeaders(session, { "content-type": "application/json" }),
      body: JSON.stringify({
        operation: "plugin.configure",
        project_id: selectedProjectID,
        plugin_id: installation.PluginID,
        version: installation.Version,
        values,
      }),
    });
    if (!response.ok) {
      setError(await response.text());
      return;
    }
    await loadAudit(session);
  }
  async function disablePlugin(installation: PluginStatus) {
    if (!selectedProjectID) {
      setError("Select an active project first");
      return;
    }
    const previewResponse = await fetch(projectAdminURL(selectedProjectID, "actions"), {
      method: "POST",
      headers: adminHeaders(session, { "content-type": "application/json" }),
      body: JSON.stringify({
        operation: "plugin.disable.preview",
        project_id: selectedProjectID,
        plugin_id: installation.PluginID,
        version: installation.Version,
      }),
    });
    if (!previewResponse.ok) {
      setError(await previewResponse.text());
      return;
    }
    const preview = (await previewResponse.json()) as { Revision: number; Dependents: string[] };
    const dependentText = preview.Dependents?.length
      ? ` Dependencies affected: ${preview.Dependents.join(", ")}.`
      : "";
    if (
      !window.confirm(
        `${text.disable} ${installation.PluginID}@${installation.Version}?${dependentText}`,
      )
    )
      return;
    const applyResponse = await fetch(projectAdminURL(selectedProjectID, "actions"), {
      method: "POST",
      headers: adminHeaders(session, { "content-type": "application/json" }),
      body: JSON.stringify({
        operation: "plugin.disable.apply",
        project_id: selectedProjectID,
        plugin_id: installation.PluginID,
        version: installation.Version,
        revision: preview.Revision,
        dependents: preview.Dependents,
        confirmed: true,
      }),
    });
    if (!applyResponse.ok) {
      setError(await applyResponse.text());
      return;
    }
    await loadPlugins(session, selectedProjectID);
    await loadAudit(session);
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
    return (
      <main className="shell" dir={language === "fa" ? "rtl" : "ltr"}>
        <section className="card login">
          <p className="eyebrow">{text.eyebrow}</p>
          <p className="muted">Loading operator session…</p>
        </section>
      </main>
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
  if (namespace && !selectedProjectID) return <p role="alert">Project unavailable. <Link href={systemUIURL("projects")}>Choose a project</Link></p>;
  if (!canViewAdminPage(activePage, activePluginIDs))
    return (
      <main className="shell" dir={language === "fa" ? "rtl" : "ltr"}>
        <section className="card admin-content" id="admin-content">
          <AdminHeader title={pageTitle} eyebrow={text.eyebrow} />
          <p className="error">This page is unavailable until its plugin is ready and enabled.</p>
          <Link href={selectedProjectID ? projectUIURL(selectedProjectID, "plugins") : systemUIURL("projects")}>Open Plugins to enable it</Link>
          <AdminFooter />
        </section>
      </main>
    );
  const pageContent = (
    <>
      <ProjectsPanel
        activePage={activePage}
        text={text}
        selectedProjectID={selectedProjectID}
        projects={projects}
        projectID={projectID}
        setProjectID={setProjectID}
        projectName={projectName}
        setProjectName={setProjectName}
        projectSlug={projectSlug}
        setProjectSlug={setProjectSlug}
        configureProject={configureProject}
        createProject={createProject}
        setProjectActive={setProjectActive}
        uninstallProject={uninstallProject}
        error={error}
        onSelectProject={(id) => {
          selectProject(id);
          void Promise.all([loadPlugins(session, id), loadFunctionDrafts(session, id)]).catch(reason => {
            if (projectRequests.current() === id) setError(String(reason));
          });
        }}
      />
      <PluginCatalogPanel
        activePage={activePage}
        projectID={selectedProjectID}
        text={text}
        manifests={manifests}
        installations={installations}
        pluginFilter={pluginFilter}
        setPluginFilter={setPluginFilter}
        installPlugin={installPlugin}
        runPluginAction={runPluginAction}
      />
      <AuditPanel
        activePage={activePage}
        text={text}
        audit={audit}
        auditFilter={auditFilter}
        setAuditFilter={setAuditFilter}
      />
      <SystemPanel
        activePage={activePage}
        text={text}
        system={system}
      />
      <FunctionDraftsPanel
        activePage={activePage}
        text={text}
        functionName={functionName}
        setFunctionName={setFunctionName}
        functionSource={functionSource}
        setFunctionSource={setFunctionSource}
        functionInput={functionInput}
        setFunctionInput={setFunctionInput}
        submitFunction={submitFunction}
        functionMessage={functionMessage}
        functionDrafts={functionDrafts}
        system={system}
      />
      <RoutesPanel
        activePage={activePage}
        text={text}
        playgroundRoute={playgroundRoute}
        setPlaygroundRoute={setPlaygroundRoute}
        projectRoutes={projectRoutes}
        playgroundPayload={playgroundPayload}
        setPlaygroundPayload={setPlaygroundPayload}
        testPlayground={testPlayground}
        sendPlayground={sendPlayground}
        playgroundMessage={playgroundMessage}
        selectedProjectID={selectedProjectID}
      />
      <MonitoringPanel
        activePage={activePage}
        text={text}
        manifests={manifests}
        installations={installations}
      />
      <section className="page-hidden">
        <div className="playground" />
      </section>
    </>
  );
  if (pathname !== systemUIURL()) return pageContent;
  return (
    <main className="shell" dir={language === "fa" ? "rtl" : "ltr"}>
      <a className="skip-link" href="#admin-content">Skip to content</a>
      <div className="workspace">
        <AdminSidebar
          activePage={activePage}
          projectID={selectedProjectID}
          onSignOut={() => void signOut()}
          visiblePages={visibleAdminPages(manifests.length, activePluginIDs)}
        />
        <section className="card admin-content" id="admin-content">
          <AdminHeader title={pageTitle} eyebrow={text.eyebrow} />
          {pageContent}
          <AdminFooter />
        </section>
      </div>
    </main>
  );
}
