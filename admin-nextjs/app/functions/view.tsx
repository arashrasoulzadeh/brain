"use client";
import { adminHeaders, readAdminSession, preferredAdminProject, projectAdminURL, systemAdminURL, projectUIURL, adminUIContext } from "../lib/admin-api";

import { usePathname } from "next/navigation";
import { useEffect, useState } from "react";
import Link from "next/link";
import { AdminPageFrame } from "../components/admin-chrome";

type Draft = { project_id: string; name: string; source: string; version?: number; public?: boolean; permissions?: string[]; created_at: string };

export default function FunctionsPage() {
  const { namespace } = adminUIContext(usePathname());
  const [project, setProject] = useState("");
  const [drafts, setDrafts] = useState<Draft[]>([]);
  const [error, setError] = useState("");

  useEffect(() => {
    const controller = new AbortController();
    const token = readAdminSession();
    setProject(""); setDrafts([]); setError("");
    if (token) void (async () => {
      const options = { headers: adminHeaders(token), signal: controller.signal };
      const response = await fetch(systemAdminURL("projects"), options);
      if (!response.ok) throw new Error("Projects unavailable");
      const active = preferredAdminProject(await response.json() as { ID: string; Slug?: string; Active: boolean }[], namespace)?.ID;
      if (!active) throw new Error("Project unavailable");
      const functions = await fetch(projectAdminURL(active, "functions"), options);
      if (!functions.ok) throw new Error("Functions unavailable");
      const loaded = await functions.json() as Draft[];
      if (!controller.signal.aborted) { setProject(active); setDrafts(loaded.filter(draft => draft.project_id === active)); }
    })().catch(reason => { if (!controller.signal.aborted) setError(String(reason)); });
    return () => controller.abort();
  }, [namespace]);

  if (!project) return <AdminPageFrame activePage="functions" title="Functions" requiredPluginID="functions"><p role={error ? "alert" : "status"}>{error || "Loading project…"}</p></AdminPageFrame>;
  return <AdminPageFrame activePage="functions" title="Functions" projectID={project} requiredPluginID="functions"><div className="section-heading"><Link className="button-link" href={projectUIURL(project, "functions", "new")}>Create function</Link></div><p className="muted">Project-scoped serverless functions. Select a function to edit permissions, versions, analytics, and logs.</p><div className="table-wrap"><table><thead><tr><th>Function</th><th>Version</th><th>Visibility</th><th>Updated</th><th /></tr></thead><tbody>{drafts.map((draft) => <tr key={`${draft.project_id}:${draft.name}`}><td><Link href={projectUIURL(project, "functions", draft.name)}>{draft.name}</Link></td><td>{draft.version ?? 1}</td><td>{draft.public ? "Public" : "Private"}</td><td>{new Date(draft.created_at).toLocaleString()}</td><td><Link href={projectUIURL(project, "functions", draft.name)}>View</Link></td></tr>)}</tbody></table>{drafts.length === 0 && <p className="muted">No functions yet.</p>}</div></AdminPageFrame>;
}
