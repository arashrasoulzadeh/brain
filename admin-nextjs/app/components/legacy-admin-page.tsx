"use client";
import { usePathname } from "next/navigation";
import { useEffect, useState } from "react";
import Link from "next/link";
import { AdminPageFrame } from "./admin-chrome";
import { AdminPageRenderer } from "./admin-renderer";
import { adminHeaders, readAdminSession, preferredAdminProject } from "../lib/admin-api";
import { adminContributions, adminPageExtensions } from "../lib/admin-registry";
import type { AdminCatalog, AdminContribution } from "../lib/admin-registry";

export default function LegacyAdminPage() {
  const route = usePathname().split("/")[1];
  const [catalog, setCatalog] = useState<{ route: string; project: string; contributions: AdminContribution[] }>();
  const [error, setError] = useState("");
  useEffect(() => {
    const controller = new AbortController();
    setCatalog(undefined); setError("");
    void (async () => {
      const token = readAdminSession();
      if (!token) throw new Error("Sign in to continue");
      const headers = adminHeaders(token), signal = controller.signal;
      const response = await fetch("/api/admin/projects", { headers, signal });
      if (!response.ok) throw new Error("Projects unavailable");
      const projects = await response.json() as { ID: string; Active: boolean }[];
      const project = preferredAdminProject(projects)?.ID;
      if (!project) throw new Error("Select an active project");
      const plugins = await fetch("/api/admin/plugins?project_id=" + encodeURIComponent(project), { headers, signal });
      if (!plugins.ok) throw new Error("Plugin catalog unavailable");
      const data = await plugins.json() as AdminCatalog;
      if (!signal.aborted) setCatalog({ route, project, contributions: adminContributions(project, data) });
    })().catch(reason => { if (!controller.signal.aborted) setError(String(reason)); });
    return () => controller.abort();
  }, [route]);
  const current = catalog?.route === route ? catalog : undefined;
  const pages = current?.contributions.filter(item => item.page.legacy_route === route) ?? [];
  const title = route ? route[0].toUpperCase() + route.slice(1).replaceAll("-", " ") : "Plugin administration";
  return <AdminPageFrame activePage={route} title={title}>
    {error ? <><p role="alert">{error}</p><Link href="/projects">Choose a project</Link></> : !current ? <p>Loading plugin pages…</p> : !pages.length ? <><p>Page unavailable. Enable a plugin contributing this page and grant its required capabilities.</p><Link href="/plugins">Manage plugins</Link></> : pages.map(item => <AdminPageRenderer key={item.href} contribution={item} extensions={adminPageExtensions(item, current.contributions)} />)}
  </AdminPageFrame>;
}
