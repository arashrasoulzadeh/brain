"use client";

import { usePathname } from "next/navigation";
import { useEffect, useState } from "react";
import Link from "next/link";
import { AdminPageFrame } from "../components/admin-chrome";
import { AdminProgress } from "../components/admin-dialog";
import { PluginManageLink } from "../components/plugin-manage-link";
import { adminHeaders, readAdminSession, preferredAdminProject, projectAdminURL, systemAdminURL, projectUIURL, adminUIContext, ADMIN_SESSION_CHANGED, ADMIN_PROJECT_CHANGED } from "../lib/admin-api";
import { readyAdminInstallations, type AdminCatalog } from "../lib/admin-registry";
import { projectFunctionRecords, functionCatalogSummary, functionCatalogRows, functionCatalogPage, type FunctionCatalogRecord, type FunctionCatalogFilters, type FunctionCatalogSort } from "../lib/function-inspection";

type Catalog = { project: string; versions: string[]; records: FunctionCatalogRecord[] };
const defaultFilters: FunctionCatalogFilters = { query: "", visibility: "", publication: "", method: "", sort: "name" };
const pageSize = 20;

export default function FunctionsPage() {
  const { namespace } = adminUIContext(usePathname());
  const [sessionEpoch, setSessionEpoch] = useState(0);
  useEffect(() => {
    const changed = () => setSessionEpoch(value => value + 1);
    window.addEventListener(ADMIN_SESSION_CHANGED, changed);
    window.addEventListener(ADMIN_PROJECT_CHANGED, changed);
    return () => {
      window.removeEventListener(ADMIN_SESSION_CHANGED, changed);
      window.removeEventListener(ADMIN_PROJECT_CHANGED, changed);
    };
  }, []);
  const scope = (namespace ?? "") + ":" + sessionEpoch;
  return <AdminPageFrame key={scope} activePage="functions" title="Functions" requiredPluginID="functions">
    <FunctionCatalog key={scope} namespace={namespace} />
  </AdminPageFrame>;
}

function FunctionCatalog({ namespace }: { namespace?: string }) {
  const [catalog, setCatalog] = useState<Catalog>();
  const [error, setError] = useState("");
  const [refresh, setRefresh] = useState(0);
  const [filters, setFilters] = useState(defaultFilters);
  const [requestedPage, setRequestedPage] = useState(0);

  useEffect(() => {
    const controller = new AbortController();
    const token = readAdminSession();
    const current = () => !controller.signal.aborted && readAdminSession() === token;
    setCatalog(undefined);
    setError("");
    void (async () => {
      if (!token || !namespace) throw new Error("Select an active project and sign in.");
      const options = { headers: adminHeaders(token), signal: controller.signal };
      const projects = await fetch(systemAdminURL("projects"), options);
      if (!projects.ok) throw new Error("Projects unavailable (" + projects.status + ").");
      const project = preferredAdminProject(await projects.json() as { ID: string; Slug?: string; Active: boolean }[], namespace)?.ID;
      if (!project) throw new Error("Project unavailable.");
      const plugins = await fetch(projectAdminURL(project, "plugins"), options);
      if (!plugins.ok) throw new Error("Plugin catalog unavailable (" + plugins.status + ").");
      const versions = readyAdminInstallations(project, await plugins.json() as AdminCatalog, "functions").map(item => item.Version);
      if (!versions.length) throw new Error("Functions installation is not ready or lacks required capabilities.");
      if (!current()) return;
      const response = await fetch(projectAdminURL(project, "functions"), options);
      if (!response.ok) throw new Error("Functions unavailable (" + response.status + ").");
      const records = projectFunctionRecords((await response.json() as FunctionCatalogRecord[] | null) ?? [], project);
      if (current()) setCatalog({ project, versions, records });
    })().catch(reason => {
      if (current()) setError(reason instanceof Error ? reason.message : String(reason));
    });
    return () => controller.abort();
  }, [namespace, refresh]);

  function reload() {
    setCatalog(undefined);
    setError("");
    setRefresh(value => value + 1);
  }

  function updateFilters(patch: Partial<FunctionCatalogFilters>) {
    setFilters(previous => ({ ...previous, ...patch }));
    setRequestedPage(0);
  }

  function resetFilters() {
    setFilters(defaultFilters);
    setRequestedPage(0);
  }

  if (error) return <section className="section"><h2>Unable to load functions</h2><p role="alert">{error}</p><button type="button" className="secondary" onClick={reload}>Retry</button></section>;
  if (!catalog) return <AdminProgress label="Loading function catalog…" />;

  const summary = functionCatalogSummary(catalog.records);
  const filtered = functionCatalogRows(catalog.records, filters);
  const page = functionCatalogPage(filtered, requestedPage, pageSize);
  const hasFilters = Boolean(filters.query || filters.visibility || filters.publication || filters.method || filters.sort !== "name");

  return <div className="function-catalog">
    <section className="section" aria-label="Function overview">
      <div className="section-heading">
        <div><h2>Function catalog</h2><p className="muted">Project-scoped serverless functions · {catalog.project}</p></div>
        <div className="function-catalog-actions">
          <button type="button" className="secondary" onClick={reload}>Refresh functions</button>
          <Link className="button-link" href={projectUIURL(catalog.project, "functions", "new")}>Create function</Link>
        </div>
      </div>
      <div className="metric-grid function-catalog-metrics">
        {([["All functions", summary.total], ["Published records", summary.published], ["Unpublished", summary.unpublished], ["Public visibility", summary.public]] as const).map(([label, count]) =>
          <article key={label}><span>{label}</span><strong>{count}</strong></article>)}
      </div>
      <p className="muted">Counts include only functions you can inspect. Publication and configured visibility do not guarantee that a caller is authorized.</p>
    </section>
    <section className="section" aria-labelledby="function-directory-title">
      <div className="section-heading"><h2 id="function-directory-title">Directory</h2><span className="muted" role="status">{filtered.length} / {summary.total} functions</span></div>
      <div className="function-catalog-filters">
        <label>Search functions<input type="search" placeholder="Search by name" value={filters.query} onChange={event => updateFilters({ query: event.target.value })} /></label>
        <label>Publication<select value={filters.publication} onChange={event => updateFilters({ publication: event.target.value })}><option value="">All publication states</option><option value="published">Published</option><option value="draft">Unpublished</option></select></label>
        <label>Visibility<select value={filters.visibility} onChange={event => updateFilters({ visibility: event.target.value })}><option value="">All visibility</option><option value="public">Public</option><option value="private">Private</option></select></label>
        <label>HTTP method<select value={filters.method} onChange={event => updateFilters({ method: event.target.value })}><option value="">All methods</option>{["GET", "POST", "PUT", "PATCH"].map(method => <option key={method}>{method}</option>)}</select></label>
        <label>Sort by<select value={filters.sort} onChange={event => updateFilters({ sort: event.target.value as FunctionCatalogSort })}><option value="name">Name (A–Z)</option><option value="newest">Newest created</option><option value="published">Latest publication</option><option value="version">Highest version</option></select></label>
        <button type="button" className="secondary" disabled={!hasFilters} onClick={resetFilters}>Reset filters</button>
      </div>
      {!summary.total ? <div className="empty-panel"><h3>No functions yet</h3><p>No function records are available to inspect in this project.</p><Link className="button-link" href={projectUIURL(catalog.project, "functions", "new")}>Create function</Link></div>
        : !filtered.length ? <div className="empty-panel"><h3>No functions match the filters</h3><p>Try a different search or reset your filters.</p><button type="button" className="secondary" onClick={resetFilters}>Reset filters</button></div>
        : <>
          <div className="table-wrap"><table className="function-catalog-table">
            <thead><tr><th scope="col">Function</th><th scope="col">Method</th><th scope="col">Version</th><th scope="col">Publication</th><th scope="col">Visibility</th><th scope="col">Permissions</th><th scope="col">Created</th><th scope="col">Last published</th></tr></thead>
            <tbody>{page.rows.map(record => <tr key={record.name}>
              <th scope="row"><Link href={projectUIURL(catalog.project, "functions", record.name)}>{record.name}</Link></th>
              <td><code>{record.method || "—"}</code></td><td>v{record.version ?? 1}</td>
              <td><span className={"status-pill " + (record.published_at ? "is-success" : "is-inactive")}>{record.published_at ? "Published" : "Unpublished"}</span></td>
              <td>{record.public ? "Public" : "Private"}</td><td>{record.permissions?.length ?? 0}</td>
              <td><CatalogDate value={record.created_at} /></td><td><CatalogDate value={record.published_at} /></td>
            </tr>)}</tbody>
          </table></div>
          <nav className="function-catalog-pagination" aria-label="Function catalog pagination">
            <span className="muted">Showing {page.start}–{page.end} of {filtered.length}</span>
            <div className="function-catalog-actions"><button type="button" className="secondary" disabled={page.page === 0} onClick={() => setRequestedPage(page.page - 1)}>Previous</button><span>Page {page.page + 1} of {page.pages}</span><button type="button" className="secondary" disabled={page.page + 1 === page.pages} onClick={() => setRequestedPage(page.page + 1)}>Next</button></div>
          </nav>
        </>}
    </section>
    <section className="section" aria-label="Functions installation">
      <div className="section-heading"><div><h2>Functions installation</h2><p className="muted">Manage the project plugin separately from function source versions.</p></div>
        <div className="function-catalog-actions">{catalog.versions.map(version => <div key={version} className="function-catalog-actions"><span>Plugin v{version}</span><PluginManageLink projectID={catalog.project} pluginID="functions" version={version} /></div>)}</div>
      </div>
    </section>
  </div>;
}

function CatalogDate({ value }: { value?: string }) {
  if (!value || !Number.isFinite(Date.parse(value))) return <>—</>;
  return <time dateTime={value} title={new Date(value).toLocaleString()}>{new Date(value).toLocaleDateString()}</time>;
}
