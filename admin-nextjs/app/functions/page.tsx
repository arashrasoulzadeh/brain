"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { AdminPageFrame } from "../components/admin-chrome";

type Draft = { project_id: string; name: string; source: string; version?: number; public?: boolean; permissions?: string[]; created_at: string };

export default function FunctionsPage() {
  const [session, setSession] = useState("");
  const [project, setProject] = useState("");
  const [drafts, setDrafts] = useState<Draft[]>([]);

  useEffect(() => {
    const token = window.sessionStorage.getItem("brain-admin-session") ?? "";
    setSession(token);
    if (!token) return;
    void (async () => { const projects = await (await fetch("/api/admin/projects", { headers: { authorization: `Bearer ${token}` } })).json() as { ID: string; Active: boolean }[]; const active = projects.find((candidate) => candidate.Active)?.ID ?? ""; setProject(active); if (active) setDrafts(await (await fetch(`/api/admin/functions?project_id=${encodeURIComponent(active)}`, { headers: { authorization: `Bearer ${token}` } })).json() as Draft[]); })();
  }, []);

  return <AdminPageFrame activePage="functions" title="Functions" projectID={project} requiredPluginID="functions"><div className="section-heading"><Link className="button-link" href="/functions/new">Create function</Link></div><p className="muted">Project-scoped serverless functions. Select a function to edit permissions, versions, analytics, and logs.</p><div className="table-wrap"><table><thead><tr><th>Function</th><th>Version</th><th>Visibility</th><th>Updated</th><th /></tr></thead><tbody>{drafts.map((draft) => <tr key={`${draft.project_id}:${draft.name}`}><td><Link href={`/functions/${encodeURIComponent(draft.name)}`}>{draft.name}</Link></td><td>{draft.version ?? 1}</td><td>{draft.public ? "Public" : "Private"}</td><td>{new Date(draft.created_at).toLocaleString()}</td><td><Link href={`/functions/${encodeURIComponent(draft.name)}`}>View</Link></td></tr>)}</tbody></table>{drafts.length === 0 && <p className="muted">No functions yet.</p>}</div></AdminPageFrame>;
}
