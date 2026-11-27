"use client";
import { adminHeaders, readAdminSession, preferredAdminProject, projectAdminURL, systemAdminURL, systemUIURL, projectUIURL, adminUIContext } from "../lib/admin-api";
import { enabledPluginIDs, visibleAdminPages } from "../lib/admin-navigation";
import { adminContributions, adminWorkspaceFallbacks, readyAdminInstallations } from "../lib/admin-registry";
import type { AdminCatalog, AdminContribution } from "../lib/admin-registry";

import Link from "next/link";
import type { ReactNode } from "react";
import { createContext, useContext, useEffect, useState } from "react";
import { adminMenu, menuItemActive, type MenuIcon, type MenuItem } from "../lib/admin-menu";
import { usePathname, useRouter } from "next/navigation";
import { ADMIN_PROJECT_CHANGED, ADMIN_PLUGIN_CATALOG_CHANGED } from "../lib/admin-api";

type AdminSidebarProps = {
  activePage: string;
  projectID?: string;
  onSignOut?: () => void;
  title?: string;
  visiblePages?: string[];
};

export function AdminSidebar({
  activePage,
  projectID,
  onSignOut,
  title = "Brain",
  visiblePages,
}: AdminSidebarProps) {
  const pathname = usePathname();
  const requestedProject = adminUIContext(pathname).namespace ?? (projectID || undefined);
  const [projectEpoch, setProjectEpoch] = useState(0);
  const [activeProject, setActiveProject] = useState("");
  const [loadedScope, setLoadedScope] = useState<string>();
  const [search, setSearch] = useState("");
  const [mobileOpen, setMobileOpen] = useState(false);
  const [collapsed, setCollapsed] = useState<Record<string, boolean>>({});
  useEffect(() => {
    const changed = () => setProjectEpoch(value => value + 1);
    window.addEventListener(ADMIN_PROJECT_CHANGED, changed);
    window.addEventListener(ADMIN_PLUGIN_CATALOG_CHANGED, changed);
    return () => { window.removeEventListener(ADMIN_PROJECT_CHANGED, changed); window.removeEventListener(ADMIN_PLUGIN_CATALOG_CHANGED, changed); };
  }, []);
  useEffect(() => { setMobileOpen(false); setSearch(""); }, [pathname]);
  const [discoveredPages, setDiscoveredPages] = useState<string[]>(["projects"]);
  const [extensions, setExtensions] = useState<AdminContribution[]>([]);
  const [fallback, setFallback] = useState<MenuItem[]>([]);
  useEffect(() => {
    const controller = new AbortController();
    setExtensions([]);
    setFallback([]);
    setActiveProject("");
    setDiscoveredPages(["projects"]);
    const token = readAdminSession();
    if (!token) return () => controller.abort();
    const headers = adminHeaders(token);
    const signal = controller.signal;
    void (async () => {
      const projectsResponse = await fetch(systemAdminURL("projects"), { headers, signal });
      if (!projectsResponse.ok) return;
      const projects = (await projectsResponse.json()) as Array<{ ID: string; Slug?: string; Active: boolean }>;
      const active = preferredAdminProject(projects, requestedProject)?.ID;
      if (!active || signal.aborted) return;
      setActiveProject(active);
      setLoadedScope(requestedProject);
      const response = await fetch(projectAdminURL(active, "plugins"), { headers, signal });
      if (!response.ok) return;
      const payload = (await response.json()) as AdminCatalog;
      if (signal.aborted) return;
      setDiscoveredPages(visibleAdminPages(payload.manifests?.length ?? 0, enabledPluginIDs(payload.installations ?? [])));
      const contributions = adminContributions(active, payload);
      setExtensions(contributions);
      setFallback(adminWorkspaceFallbacks(active, payload, contributions));
    })().catch(() => undefined);
    return () => controller.abort();
  }, [requestedProject, activePage, projectEpoch]);
  const menuProject = loadedScope === requestedProject ? activeProject : "";
  const groups = adminMenu(visiblePages ?? discoveredPages, menuProject ? extensions : [], search, menuProject, menuProject ? fallback : []);
  const menuPath = requestedProject && menuProject ? pathname.replace("/" + requestedProject + "/ui", "/" + menuProject + "/ui") : pathname;
  return (
    <aside className="sidebar">
      <div className="sidebar-brand">
        <Link href={systemUIURL("projects")} className="brand-link"><span className="brand-mark" aria-hidden="true"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.7"><path d="M12 3 3 8l9 5 9-5-9-5ZM3 12l9 5 9-5M3 16l9 5 9-5" /></svg></span><span><strong>{title}</strong><small>Control plane</small></span></Link>
        <button type="button" className="menu-toggle secondary" aria-label="Toggle navigation" aria-expanded={mobileOpen} aria-controls="admin-menu" onClick={() => setMobileOpen(value => !value)}><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8"><path d="M4 6h16M4 12h16M4 18h16" /></svg></button>
      </div>
      <div id="admin-menu" className={"sidebar-menu" + (mobileOpen ? " is-open" : "")}>
        <Link href={systemUIURL("projects")} className="workspace-picker" title="Change project"><MenuGlyph name="projects" /><span><small>Current project</small><strong>{menuProject || "Select project"}</strong></span><span aria-hidden="true">⌄</span></Link>
        <label className="menu-search"><MenuGlyph name="search" /><input aria-label="Filter navigation" placeholder="Find a page…" value={search} onChange={event => setSearch(event.target.value)} /></label>
        <nav className="sidebar-nav" aria-label="Admin navigation">
          {groups.map(group => {
            const active = group.items.some(item => menuItemActive(menuPath, item.href));
            const filtering = !!search.trim();
            const expanded = filtering || !(collapsed[group.label] ?? (group.label === "Extensions" && !active));
            return <div key={group.label} className="menu-group">
              <button type="button" className="menu-group-title" aria-expanded={expanded} disabled={filtering} onClick={() => setCollapsed(previous => ({ ...previous, [group.label]: expanded }))}>{group.label}<span aria-hidden="true">{expanded ? "−" : "+"}</span></button>
              {expanded && group.items.map(item => {
                const current = menuItemActive(menuPath, item.href);
                return <Link className={current ? "active" : ""} aria-current={current ? "page" : undefined} href={item.href} title={item.detail ? item.label + " · " + item.detail : item.label} key={item.href}><MenuGlyph name={item.icon} /><span className="menu-label">{item.label}{item.detail && <small>{item.detail}</small>}</span></Link>;
              })}
            </div>;
          })}
          {!groups.length && <p className="menu-empty">No matching pages.</p>}
        </nav>
        {onSignOut && <div className="sidebar-actions"><button type="button" className="secondary" onClick={onSignOut}><MenuGlyph name="logout" />Sign out</button></div>}
      </div>
    </aside>
  );
}

const glyphs: Record<MenuIcon | "search" | "logout", string> = {
  projects: "M3 7h7l2 2h9v10H3z M3 7V5h7l2 2",
  plugins: "M8 3h8v5h5v8h-5v5H8v-5H3V8h5z",
  audit: "M6 3h12v18H6z M9 8h6 M9 12h6 M9 16h4",
  system: "M12 3l8 5v8l-8 5-8-5V8z M12 8v8 M8 12h8",
  functions: "M8 5H5v14h3 M16 5h3v14h-3 M14 8l-4 8",
  routes: "M5 5h6v6H5z M13 13h6v6h-6z M11 8h5v5 M8 11v5h5",
  monitoring: "M3 19V5 M3 19h18 M6 14l4-5 4 3 6-7",
  workers: "M4 5h16v6H4z M4 13h16v6H4z M7 8h.1 M7 16h.1",
  resource: "M4 6c0-2 16-2 16 0s-16 2-16 0z M4 6v12c0 2 16 2 16 0V6 M4 12c0 2 16 2 16 0",
  extension: "M12 3l9 9-9 9-9-9z M9 12h6 M12 9v6",
  search: "M10 4a6 6 0 1 0 0 12 6 6 0 0 0 0-12z M15 15l5 5",
  logout: "M10 4H4v16h6 M10 12h11 M17 8l4 4-4 4",
};

function MenuGlyph({ name }: { name: keyof typeof glyphs }) {
  return <svg className="menu-glyph" aria-hidden="true" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.65" strokeLinecap="round" strokeLinejoin="round"><path d={glyphs[name]} /></svg>;
}

export type PageHeading = { path: string; title: string; eyebrow?: string };
export const AdminHeadingContext = createContext<((heading: PageHeading) => void) | null>(null);

export function AdminHeader({
  title,
  eyebrow = "BRAIN CONTROL PLANE",
}: {
  title: string;
  eyebrow?: string;
}) {
  return (
    <header className="admin-header">
      <p className="eyebrow">{eyebrow}</p>
      <h1>{title}</h1>
    </header>
  );
}

export function AdminFooter() {
  return <footer className="admin-footer">Brain Control Plane · Local administration</footer>;
}

export function AdminPageFrame({
  activePage,
  title,
  eyebrow,
  projectID,
  onSignOut,
  requiredPluginID,
  children,
}: {
  activePage: string;
  title: string;
  eyebrow?: string;
  projectID?: string;
  onSignOut?: () => void;
  requiredPluginID?: string;
  children: ReactNode;
}) {
  const router = useRouter();
  const pathname = usePathname();
  const setHeading = useContext(AdminHeadingContext);
  useEffect(() => { setHeading?.({ path: pathname, title, eyebrow }); }, [setHeading, pathname, title, eyebrow]);
  const [access, setAccess] = useState<{ path: string; epoch: number; status: "allowed" | "denied"; message?: string; project?: string }>();
  const [catalogEpoch, setCatalogEpoch] = useState(0);
  useEffect(() => {
    const changed = () => setCatalogEpoch(value => value + 1);
    window.addEventListener(ADMIN_PLUGIN_CATALOG_CHANGED, changed);
    return () => window.removeEventListener(ADMIN_PLUGIN_CATALOG_CHANGED, changed);
  }, []);
  useEffect(() => {
    if (!requiredPluginID) return;
    const controller = new AbortController();
    const token = readAdminSession();
    if (!token) { router.replace(systemUIURL()); return; }
    const options = { headers: adminHeaders(token), signal: controller.signal };
    let project: string | undefined;
    void (async () => {
      const response = await fetch(systemAdminURL("projects"), options);
      if (!response.ok) throw new Error("Project access unavailable");
      project = preferredAdminProject(await response.json() as Array<{ ID: string; Slug?: string; Active: boolean }>, adminUIContext(pathname).namespace)?.ID;
      if (!project) throw new Error("Project unavailable");
      const plugins = await fetch(projectAdminURL(project, "plugins"), options);
      if (!plugins.ok) throw new Error("Plugin catalog unavailable");
      const catalog = await plugins.json() as AdminCatalog;
      const allowed = readyAdminInstallations(project, catalog, requiredPluginID).length > 0;
      if (!allowed) throw new Error("This plugin is disabled, draining, not ready or missing required capabilities.");
      if (!controller.signal.aborted && readAdminSession() === token) setAccess({ path: pathname, epoch: catalogEpoch, status: "allowed", project });
    })().catch(reason => {
      if (!controller.signal.aborted) setAccess({ path: pathname, epoch: catalogEpoch, status: "denied", message: String(reason), project });
    });
    return () => controller.abort();
  }, [requiredPluginID, router, pathname, catalogEpoch]);
  if (requiredPluginID) {
    if (!access || access.path !== pathname || access.epoch !== catalogEpoch) return <p className="muted" role="status">Checking page access…</p>;
    if (access.status === "denied") return <section className="section"><p role="alert">{access.message}</p><Link href={access.project ? projectUIURL(access.project, "plugins") : systemUIURL("projects")}>{access.project ? "Manage plugins" : "Choose a project"}</Link><button type="button" className="secondary" onClick={() => setCatalogEpoch(value => value + 1)}>Retry</button></section>;
  }
  return <>{children}</>;
}
