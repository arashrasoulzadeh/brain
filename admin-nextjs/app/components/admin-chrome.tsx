"use client";
import { adminHeaders, readAdminSession } from "../lib/admin-api";
import { enabledPluginIDs, visibleAdminPages } from "../lib/admin-navigation";
import { adminContributions } from "../lib/admin-registry";
import type { AdminCatalog, AdminContribution } from "../lib/admin-registry";

import Link from "next/link";
import type { ReactNode } from "react";
import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";

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
  title = "Projects",
  visiblePages,
}: AdminSidebarProps) {
  const links = [
    ["projects", "Projects"],
    ["plugins", "Plugins"],
    ["system", "System capabilities"],
    ["audit", "Audit log"],
    ["database", "Database"],
    ["functions", "Functions"],
    ["workflows", "Workflows"],
    ["files", "Files"],
    ["live-sessions", "Live sessions"],
    ["routes", "Routes"],
    ["monitoring", "Monitoring"],
    ["workers", "Workers"],
  ] as const;
  const params = useParams<{ project?: string }>();
  const requestedProject = projectID ?? params.project;
  const [discoveredPages, setDiscoveredPages] = useState<string[]>(["projects"]);
  const [extensions, setExtensions] = useState<AdminContribution[]>([]);
  useEffect(() => {
    const controller = new AbortController();
    setExtensions([]);
    setDiscoveredPages(["projects"]);
    const token = readAdminSession();
    if (!token) return () => controller.abort();
    const headers = adminHeaders(token);
    const signal = controller.signal;
    void (async () => {
      const projectsResponse = await fetch("/api/admin/projects", { headers, signal });
      if (!projectsResponse.ok) return;
      const projects = (await projectsResponse.json()) as Array<{ ID: string; Active: boolean }>;
      const active = requestedProject === undefined
        ? projects.find(candidate => candidate.Active)?.ID
        : projects.find(candidate => candidate.Active && candidate.ID === requestedProject)?.ID;
      if (!active) return;
      const response = await fetch("/api/admin/plugins?project_id=" + encodeURIComponent(active), { headers, signal });
      if (!response.ok) return;
      const payload = (await response.json()) as AdminCatalog;
      if (signal.aborted) return;
      setDiscoveredPages(visibleAdminPages(payload.manifests?.length ?? 0, enabledPluginIDs(payload.installations ?? [])));
      setExtensions(adminContributions(active, payload));
    })().catch(() => undefined);
    return () => controller.abort();
  }, [requestedProject, activePage]);
  const allowedPages = visiblePages ?? discoveredPages;
  return (
    <aside className="sidebar">
      <p className="eyebrow">BRAIN CONTROL PLANE</p>
      <h2>{title}</h2>
      <nav className="sidebar-nav" aria-label="Admin navigation">
        {links
          .filter(([href]) => allowedPages.includes(href))
          .map(([href, label]) => (
            <Link className={activePage === href ? "active" : ""} href={`/${href}`} key={href}>
              {label}
            </Link>
          ))}
        {extensions.map(item => (
          <Link href={item.href} key={item.href}>{item.page.title} · {item.pluginID}@{item.version}</Link>
        ))}
      </nav>
      {projectID !== undefined && (
        <label className="sidebar-project">
          Project
          <select value={projectID} disabled>
            <option>{projectID || "—"}</option>
          </select>
        </label>
      )}
      {onSignOut && (
        <div className="sidebar-actions">
          <button className="secondary" onClick={onSignOut}>
            Sign out
          </button>
        </div>
      )}
    </aside>
  );
}

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
  const [access, setAccess] = useState(requiredPluginID ? "checking" : "allowed");
  useEffect(() => {
    if (!requiredPluginID) return;
    const token = readAdminSession();
    if (!token) {
      router.replace("/");
      return;
    }
    void fetch("/api/admin/projects", { headers: adminHeaders(token) })
      .then(async (projectsResponse) => {
        if (!projectsResponse.ok) throw new Error("session expired");
        const projects = (await projectsResponse.json()) as Array<{ ID: string; Active: boolean }>;
        const activeProject = projects.find((project) => project.Active)?.ID;
        if (!activeProject) throw new Error("no active project");
        const pluginsResponse = await fetch(
          `/api/admin/plugins?project_id=${encodeURIComponent(activeProject)}`,
          { headers: adminHeaders(token) },
        );
        if (!pluginsResponse.ok) throw new Error("plugin access unavailable");
        const payload = (await pluginsResponse.json()) as {
          installations: Array<{ PluginID: string; Enabled: boolean; Ready: boolean }>;
        };
        const active = payload.installations.some(
          (installation) =>
            installation.PluginID === requiredPluginID &&
            installation.Enabled &&
            installation.Ready,
        );
        if (!active) {
          router.replace("/");
          return;
        }
        setAccess("allowed");
      })
      .catch(() => router.replace("/"));
  }, [requiredPluginID, router]);
  if (access !== "allowed") return <p className="muted">Checking page access…</p>;
  return <>{children}</>;
}
