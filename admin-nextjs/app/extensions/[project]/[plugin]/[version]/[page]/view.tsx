"use client";
import { useParams } from "next/navigation";
import { useEffect, useState } from "react";
import { AdminPageFrame } from "../../../../../components/admin-chrome";
import { adminHeaders, readAdminSession, projectAdminURL, systemAdminURL, preferredAdminProject } from "../../../../../lib/admin-api";
import { adminContributions, adminPageExtensions } from "../../../../../lib/admin-registry";
import type { AdminCatalog, AdminContribution } from "../../../../../lib/admin-registry";
import { AdminPageRenderer } from "../../../../../components/admin-renderer";

export default function ExtensionPage() {
  const { project, plugin, version, page } = useParams<{ project: string; plugin: string; version: string; page: string }>();
  const [resolvedProject, setResolvedProject] = useState("");
  const [contribution, setContribution] = useState<AdminContribution>();
  const [extensions, setExtensions] = useState<AdminContribution[]>([]);
  const [error, setError] = useState("");
  useEffect(() => {
    const controller = new AbortController();
    setResolvedProject(""); setContribution(undefined); setExtensions([]); setError("");
    void (async () => {
      const token = readAdminSession();
      if (!token) throw new Error("Sign in to continue");
      const headers = adminHeaders(token), signal = controller.signal;
      const projects = await fetch(systemAdminURL("projects"), { headers, signal });
      if (!projects.ok) throw new Error("Project unavailable");
      const active = preferredAdminProject(await projects.json() as Array<{ ID: string; Slug?: string; Active: boolean }>, project);
      if (!active) throw new Error("Project unavailable");
      const response = await fetch(projectAdminURL(active.ID, "plugins"), { headers, signal });
      if (!response.ok) throw new Error("Plugin catalog unavailable");
      const catalog = await response.json() as AdminCatalog;
      const contributions = adminContributions(active.ID, catalog);
      const found = contributions.find(item => item.pluginID === plugin && item.version === version && item.page.id === page);
      if (!found) throw new Error("Page unavailable or permission revoked");
      if (!signal.aborted) { setResolvedProject(active.ID); setContribution(found); setExtensions(adminPageExtensions(found, contributions)); }
    })().catch(reason => { if (!controller.signal.aborted) setError(String(reason)); });
    return () => controller.abort();
  }, [project, plugin, version, page]);
  const current = contribution?.projectID === resolvedProject && contribution.pluginID === plugin && contribution.version === version && contribution.page.id === page ? contribution : undefined;
  return <AdminPageFrame activePage="extensions" title={current?.page.title ?? "Plugin administration"}>
    {error ? <p role="alert">{error}</p> : current ? <AdminPageRenderer contribution={current} extensions={extensions} /> : <p>Loading plugin page…</p>}
  </AdminPageFrame>;
}
