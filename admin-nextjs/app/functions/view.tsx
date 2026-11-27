"use client";
import { adminHeaders, readAdminSession, preferredAdminProject, projectAdminURL, systemAdminURL, projectUIURL, adminUIContext } from "../lib/admin-api";

import { usePathname } from "next/navigation";
import { useEffect, useState } from "react";
import Link from "next/link";
import { AdminPageFrame } from "../components/admin-chrome";
import { PluginManageLink } from "../components/plugin-manage-link";
import type { AdminCatalog } from "../lib/admin-registry";
import { filterFunctionDrafts } from "../lib/function-inspection";

type Draft = { project_id: string; name: string; source: string; version?: number; public?: boolean; permissions?: string[]; created_at: string; published_at?: string };

export default function FunctionsPage() {
  const { namespace } = adminUIContext(usePathname());
  const [project, setProject] = useState("");
  const [versions, setVersions] = useState<string[]>([]);
  const [drafts, setDrafts] = useState<Draft[]>([]);
  const [error, setError] = useState("");
  const [query, setQuery] = useState("");
  const [visibility, setVisibility] = useState("");
  const [publication, setPublication] = useState("");
  const [refresh, setRefresh] = useState(0);
  const filtered = filterFunctionDrafts(drafts, query, visibility, publication);

  useEffect(() => {
    const controller = new AbortController();
    const token = readAdminSession();
    setProject(""); setDrafts([]); setVersions([]); setError("");
    if (token) void (async () => {
      const options = { headers: adminHeaders(token), signal: controller.signal };
      const response = await fetch(systemAdminURL("projects"), options);
      if (!response.ok) throw new Error("Projects unavailable");
      const active = preferredAdminProject(await response.json() as { ID: string; Slug?: string; Active: boolean }[], namespace)?.ID;
      if (!active) throw new Error("Project unavailable");
      const functions = await fetch(projectAdminURL(active, "functions"), options);
      if (!functions.ok) throw new Error("Functions unavailable");
      const loaded = (await functions.json() as Draft[] | null) ?? [];
      const plugins = await fetch(projectAdminURL(active, "plugins"), options);
      if (!plugins.ok) throw new Error("Plugins unavailable");
      const catalog = await plugins.json() as AdminCatalog;
      if (!controller.signal.aborted && readAdminSession() === token) setVersions((catalog.installations ?? []).filter(item => item.ProjectID === active && item.PluginID === "functions" && item.Enabled && item.Ready && !item.Draining).map(item => item.Version));
      if (!controller.signal.aborted && readAdminSession() === token) { setProject(active); setDrafts(loaded.filter(draft => draft.project_id === active)); }
    })().catch(reason => { if (!controller.signal.aborted) setError(String(reason)); });
    return () => controller.abort();
  }, [namespace, refresh]);

  if (!project) return <AdminPageFrame activePage="functions" title="Functions" requiredPluginID="functions"><p role={error ? "alert" : "status"}>{error || "Loading project…"}</p>{error && <button type="button" onClick={() => setRefresh(value => value + 1)}>Retry</button>}</AdminPageFrame>;
  return <AdminPageFrame activePage="functions" title="Functions" projectID={project} requiredPluginID="functions"><div className="section-heading"><Link className="button-link" href={projectUIURL(project, "functions", "new")}>Create function</Link><div className="lifecycle-actions">{versions.map(version => <PluginManageLink key={version} projectID={project} pluginID="functions" version={version} />)}</div></div><p className="muted">Project-scoped serverless functions. Select a function to edit its source and permissions, inspect its current publication, analytics, and logs.</p><div className="lifecycle-actions"><label>Search functions<input type="search" value={query} onChange={event => setQuery(event.target.value)} /></label><label>Visibility<select value={visibility} onChange={event => setVisibility(event.target.value)}><option value="">All visibility</option><option value="public">Public</option><option value="private">Private</option></select></label><label>Publication<select value={publication} onChange={event => setPublication(event.target.value)}><option value="">All publication states</option><option value="published">Published</option><option value="draft">Unpublished</option></select></label><button type="button" className="secondary" onClick={() => setRefresh(value => value + 1)}>Refresh functions</button></div><p role="status">{filtered.length} / {drafts.length} functions</p><div className="table-wrap"><table><thead><tr><th>Function</th><th>Version</th><th>Visibility</th><th>Publication</th><th>Updated</th><th /></tr></thead><tbody>{filtered.map((draft) => <tr key={`${draft.project_id}:${draft.name}`}><td><Link href={projectUIURL(project, "functions", draft.name)}>{draft.name}</Link></td><td>{draft.version ?? 1}</td><td>{draft.public ? "Public" : "Private"}</td><td>{draft.published_at ? "Published" : "Unpublished"}</td><td>{new Date(draft.created_at).toLocaleString()}</td><td><Link href={projectUIURL(project, "functions", draft.name)}>View</Link></td></tr>)}</tbody></table>{drafts.length === 0 ? <p className="muted">No functions yet.</p> : !filtered.length && <p className="muted">No functions match the filters.</p>}</div></AdminPageFrame>;
}
