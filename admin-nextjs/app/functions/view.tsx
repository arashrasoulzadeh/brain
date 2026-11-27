"use client";
import { useEffect, useState } from "react";
import { usePathname } from "next/navigation";
import Link from "next/link";
import { adminUIContext, readAdminSession, projectUIURL, projectAdminURL, adminHeaders } from "../lib/admin-api";
import { AdminPageFrame } from "../components/admin-chrome";
import { AdminProgress } from "../components/admin-dialog";
import { PluginManageLink } from "../components/plugin-manage-link";
import type { AdminCatalog } from "../lib/admin-registry";
import { filterFunctionDrafts } from "../lib/function-inspection";
import { catalogStats, formatFunctionDate, loadFunctionDrafts, publicationURL, resolveFunctionScope, type FunctionDraft, type FunctionScope } from "../lib/function-studio";

export default function FunctionsPage() {
  const { namespace } = adminUIContext(usePathname());
  return <AdminPageFrame activePage="functions" title="Functions" requiredPluginID="functions"><FunctionCatalog key={namespace ?? ""} namespace={namespace} /></AdminPageFrame>;
}

function FunctionCatalog({ namespace }: { namespace?: string }) {
  const [scope, setScope] = useState<FunctionScope>();
  const [drafts, setDrafts] = useState<FunctionDraft[]>([]);
  const [versions, setVersions] = useState<string[]>([]);
  const [query, setQuery] = useState("");
  const [visibility, setVisibility] = useState("");
  const [publication, setPublication] = useState("");
  const [page, setPage] = useState(0);
  const [refresh, setRefresh] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  useEffect(() => {
    const controller = new AbortController();
    const token = readAdminSession();
    setLoading(true); setError(""); setDrafts([]); setScope(undefined); setVersions([]);
    void (async () => {
      const current = await resolveFunctionScope(token, namespace, controller.signal);
      const [rows, response] = await Promise.all([loadFunctionDrafts(current, controller.signal), fetch(projectAdminURL(current.project, "plugins"), { headers: adminHeaders(token), signal: controller.signal })]);
      if (!response.ok) throw new Error("Plugin catalog unavailable.");
      const catalog = await response.json() as AdminCatalog;
      if (controller.signal.aborted || readAdminSession() !== token) return;
      setScope(current); setDrafts(rows);
      setVersions((catalog.installations ?? []).filter(item => item.ProjectID === current.project && item.PluginID === "functions" && item.Enabled && item.Ready && !item.Draining).map(item => item.Version));
    })().catch(reason => { if (!controller.signal.aborted && readAdminSession() === token) setError(String(reason)); })
      .finally(() => { if (!controller.signal.aborted && readAdminSession() === token) setLoading(false); });
    return () => controller.abort();
  }, [namespace, refresh]);
  const filtered = filterFunctionDrafts(drafts, query, visibility, publication);
  const stats = catalogStats(drafts);
  const pages = Math.max(1, Math.ceil(filtered.length / 15));
  const currentPage = Math.min(page, pages - 1);
  function reset() { setQuery(""); setVisibility(""); setPublication(""); setPage(0); }
  if (loading) return <AdminProgress label="Loading functions…" />;
  if (error || !scope) return <section className="fn-panel"><p role="alert">{error || "Project unavailable."}</p><button className="secondary" type="button" onClick={() => setRefresh(value => value + 1)}>Retry</button></section>;
  return <div className="fn-workspace">
    <div className="fn-heading"><div><p className="fn-kicker">DEVELOP / SERVERLESS</p><h2>Function catalog</h2><p className="muted">Write, validate, execute and publish project-scoped JavaScript functions.</p></div><div className="fn-actions">{versions.map(version => <PluginManageLink key={version} projectID={scope.project} pluginID="functions" version={version} />)}<Link className="button-link" href={projectUIURL(scope.project, "functions", "new")}>Create function</Link></div></div>
    <div className="fn-metrics">{[["Total functions", stats.total], ["Published", stats.published], ["Unpublished", stats.unpublished], ["Public configured", stats.public]].map(([label, count]) => <article key={label}><span>{label}</span><strong>{count}</strong></article>)}</div>
    <section className="fn-panel fn-catalog">
      <div className="fn-toolbar"><label className="fn-search">Search functions<input type="search" placeholder="Find by function name…" value={query} onChange={event => { setQuery(event.target.value); setPage(0); }} /></label><label>Visibility<select value={visibility} onChange={event => { setVisibility(event.target.value); setPage(0); }}><option value="">All visibility</option><option value="public">Public</option><option value="private">Private</option></select></label><label>Publication<select value={publication} onChange={event => { setPublication(event.target.value); setPage(0); }}><option value="">All publication states</option><option value="published">Published</option><option value="draft">Unpublished</option></select></label><button type="button" className="secondary" onClick={() => setRefresh(value => value + 1)}>Refresh functions</button></div>
      <div className="fn-table-heading"><strong>Functions</strong><span role="status">{filtered.length} / {drafts.length} functions</span>{(query || visibility || publication) && <button className="secondary" type="button" onClick={reset}>Clear filters</button>}</div>
      {!drafts.length ? <div className="fn-empty"><strong>No functions yet.</strong><p className="muted">Create a function to start editing and testing code in this project.</p><Link className="button-link" href={projectUIURL(scope.project, "functions", "new")}>Create your first function</Link></div> : !filtered.length ? <div className="fn-empty"><strong>No functions match the filters.</strong><button className="secondary" type="button" onClick={reset}>Reset search</button></div> : <div className="table-wrap"><table className="fn-table"><thead><tr><th>Function</th><th>Method</th><th>Version</th><th>Visibility</th><th>Publication</th><th>Created</th><th>Published</th></tr></thead><tbody>{filtered.slice(currentPage * 15, currentPage * 15 + 15).map(draft => <tr key={draft.name}><td><Link className="fn-name" href={projectUIURL(scope.project, "functions", draft.name)}>{draft.name}</Link>{draft.published_at && <code className="fn-route-preview">{publicationURL(scope.namespace, draft.name)}</code>}</td><td><code>{draft.method || "—"}</code></td><td>v{draft.version ?? 1}</td><td><span className="fn-tag">{draft.public ? "Public" : "Private"}</span></td><td><span className={"fn-tag" + (draft.published_at ? " fn-tag-live" : "")}>{draft.published_at ? "Published" : "Unpublished"}</span></td><td>{formatFunctionDate(draft.created_at)}</td><td>{formatFunctionDate(draft.published_at)}</td></tr>)}</tbody></table></div>}
      {pages > 1 && <div className="fn-pagination"><button type="button" className="secondary" disabled={!currentPage} onClick={() => setPage(currentPage - 1)}>Previous</button><span>Page {currentPage + 1} of {pages}</span><button type="button" className="secondary" disabled={currentPage === pages - 1} onClick={() => setPage(currentPage + 1)}>Next</button></div>}
    </section>
    <p className="muted">Only functions you are authorized to inspect appear here. Configured visibility is not a guarantee of caller access.</p>
  </div>;
}
