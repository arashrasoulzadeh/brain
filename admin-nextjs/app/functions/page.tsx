"use client";
import { adminHeaders, readAdminSession, preferredAdminProject, projectAdminURL, systemAdminURL } from "../lib/admin-api";

import { useEffect, useState } from "react";
import Link from "next/link";
import { AdminPageFrame } from "../components/admin-chrome";

type Draft = { project_id: string; name: string; source: string; version?: number; public?: boolean; permissions?: string[]; created_at: string };

export default function FunctionsPage() {
  const [session, setSession] = useState("");
  const [project, setProject] = useState("");
  const [drafts, setDrafts] = useState<Draft[]>([]);

  useEffect(() => {
    const token = readAdminSession();
    setSession(token);
    if (!token) return;
    void (async () => { const projects = await (await fetch(systemAdminURL("projects"), { headers: adminHeaders(token) })).json() as { ID: string; Active: boolean }[]; const active = preferredAdminProject(projects)?.ID ?? ""; setProject(active); if (active) setDrafts(await (await fetch(projectAdminURL(active, "functions"), { headers: adminHeaders(token) })).json() as Draft[]); })();
  }, []);

  return <AdminPageFrame activePage="functions" title="Functions" projectID={project} requiredPluginID="functions"><div className="section-heading"><Link className="button-link" href="/functions/new">Create function</Link></div><p className="muted">Project-scoped serverless functions. Select a function to edit permissions, versions, analytics, and logs.</p><div className="table-wrap"><table><thead><tr><th>Function</th><th>Version</th><th>Visibility</th><th>Updated</th><th /></tr></thead><tbody>{drafts.map((draft) => <tr key={`${draft.project_id}:${draft.name}`}><td><Link href={`/functions/${encodeURIComponent(draft.name)}`}>{draft.name}</Link></td><td>{draft.version ?? 1}</td><td>{draft.public ? "Public" : "Private"}</td><td>{new Date(draft.created_at).toLocaleString()}</td><td><Link href={`/functions/${encodeURIComponent(draft.name)}`}>View</Link></td></tr>)}</tbody></table>{drafts.length === 0 && <p className="muted">No functions yet.</p>}</div></AdminPageFrame>;
}
