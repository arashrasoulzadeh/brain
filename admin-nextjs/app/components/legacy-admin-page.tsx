"use client";
import { usePathname } from "next/navigation";
import { useEffect, useState } from "react";
import Link from "next/link";
import { AdminPageFrame } from "./admin-chrome";
import { AdminPageRenderer } from "./admin-renderer";
import { adminHeaders, readAdminSession, preferredAdminProject, projectAdminURL, systemAdminURL, systemUIURL, projectUIURL, adminUIContext, ADMIN_PROJECT_CHANGED, ADMIN_PLUGIN_CATALOG_CHANGED } from "../lib/admin-api";
import { adminContributions, adminPageExtensions, adminLegacyRootPages, adminAliasMatches, adminAliasTitle } from "../lib/admin-registry";
import type { AdminCatalog, AdminContribution } from "../lib/admin-registry";

export default function LegacyAdminPage() {
  const { page: route, namespace } = adminUIContext(usePathname());
  const [catalog, setCatalog] = useState<{ route: string; project: string; contributions: AdminContribution[] }>();
  const [error, setError] = useState("");
  const [epoch, setEpoch] = useState(0);
  useEffect(() => {
    const changed = () => { setCatalog(undefined); setEpoch(value => value + 1); };
    window.addEventListener(ADMIN_PROJECT_CHANGED, changed);
    window.addEventListener(ADMIN_PLUGIN_CATALOG_CHANGED, changed);
    return () => { window.removeEventListener(ADMIN_PROJECT_CHANGED, changed); window.removeEventListener(ADMIN_PLUGIN_CATALOG_CHANGED, changed); };
  }, []);
  useEffect(() => {
    const controller = new AbortController();
    setCatalog(undefined); setError("");
    void (async () => {
      const token = readAdminSession();
      if (!token) throw new Error("Sign in to continue");
      const headers = adminHeaders(token), signal = controller.signal;
      const response = await fetch(systemAdminURL("projects"), { headers, signal });
      if (!response.ok) throw new Error("Projects unavailable");
      const projects = await response.json() as { ID: string; Slug?: string; Active: boolean }[];
      const project = preferredAdminProject(projects, namespace)?.ID;
      if (!project) throw new Error("Select an active project");
      const plugins = await fetch(projectAdminURL(project, "plugins"), { headers, signal });
      if (!plugins.ok) throw new Error("Plugin catalog unavailable");
      const data = await plugins.json() as AdminCatalog;
      if (!signal.aborted) setCatalog({ route, project, contributions: adminContributions(project, data) });
    })().catch(reason => { if (!controller.signal.aborted) setError(String(reason)); });
    return () => controller.abort();
  }, [route, namespace, epoch]);
  const current = catalog?.route === route ? catalog : undefined;
  const pages = adminLegacyRootPages(current?.contributions.filter(item => adminAliasMatches(item, route)) ?? []);
  const title = adminAliasTitle(route);
  return <AdminPageFrame activePage={route} title={title}>
    {error ? <><p role="alert">{error}</p><Link href={systemUIURL("projects")}>Choose a project</Link></> : !current ? <p>Loading plugin pages…</p> : !pages.length ? <><p role="alert">Page unavailable. Enable a plugin contributing this page and grant its required capabilities.</p><Link href={projectUIURL(current.project, "plugins")}>Manage plugins</Link></> : pages.map(item => <AdminPageRenderer key={item.href} contribution={item} extensions={adminPageExtensions(item, current.contributions)} />)}
  </AdminPageFrame>;
}
