"use client";
import { useParams } from "next/navigation";
import { useEffect, useRef, useState } from "react";
import { AdminPageFrame } from "../../../../../components/admin-chrome";
import { adminHeaders, readAdminSession } from "../../../../../lib/admin-api";
import { adminContributions } from "../../../../../lib/admin-registry";
import type { AdminCatalog, AdminContribution } from "../../../../../lib/admin-registry";
import { invokeAdminAction } from "../../../../../lib/admin-sdk";

function Result({ value }: { value: unknown }) {
  if (Array.isArray(value) && value.length && value.every(row => row && typeof row === "object" && !Array.isArray(row))) {
    const rows = value as Record<string, unknown>[];
    const columns = [...new Set(rows.flatMap(row => Object.keys(row)))];
    return <div style={{ overflowX: "auto" }}><table><thead><tr>{columns.map(column => <th key={column}>{column}</th>)}</tr></thead>
      <tbody>{rows.map((row, index) => <tr key={index}>{columns.map(column => <td key={column}>{JSON.stringify(row[column])}</td>)}</tr>)}</tbody></table></div>;
  }
  return <pre>{JSON.stringify(value, null, 2)}</pre>;
}

function ActionForm({ contribution }: { contribution: AdminContribution }) {
  const [operation, setOperation] = useState(contribution.page.actions[0].operation);
  const [payload, setPayload] = useState("{}");
  const [result, setResult] = useState<unknown>(undefined);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const pending = useRef<AbortController | null>(null);
  useEffect(() => () => pending.current?.abort(), []);
  async function run() {
    setError(""); setResult(undefined);
    try {
      const parsed: unknown = JSON.parse(payload);
      if (!window.confirm("Run " + operation + " in project " + contribution.projectID + "?")) return;
      setBusy(true);
      const controller = new AbortController();
      pending.current = controller;
      setResult(await invokeAdminAction(readAdminSession(), contribution, operation, parsed, controller.signal));
    } catch (reason) { setError(reason instanceof Error ? reason.message : String(reason)); }
    finally { setBusy(false); }
  }
  return <section className="section">
    <h2>{contribution.page.title}</h2><p>{contribution.projectID} · {contribution.pluginID}@{contribution.version}</p>
    <form onSubmit={event => { event.preventDefault(); if (!busy) void run(); }}>
      <label>Operation<select disabled={busy} value={operation} onChange={event => { setOperation(event.target.value); setResult(undefined); setError(""); }}>
        {contribution.page.actions.map(action => <option key={action.operation} value={action.operation}>{action.title}</option>)}
      </select></label>
      <label>JSON payload<textarea disabled={busy} rows={10} value={payload} onChange={event => setPayload(event.target.value)} /></label>
      <button disabled={busy} type="submit">{busy ? "Running…" : "Run action"}</button>
    </form>
    {error && <p role="alert">{error}</p>}
    {result !== undefined && <Result value={result} />}
  </section>;
}

export default function ExtensionPage() {
  const { project, plugin, version, page } = useParams<{ project: string; plugin: string; version: string; page: string }>();
  const [contribution, setContribution] = useState<AdminContribution>();
  const [error, setError] = useState("");
  useEffect(() => {
    const controller = new AbortController();
    setContribution(undefined); setError("");
    void (async () => {
      const token = readAdminSession();
      if (!token) throw new Error("Sign in to continue");
      const headers = adminHeaders(token), signal = controller.signal;
      const projects = await fetch("/api/admin/projects", { headers, signal });
      if (!projects.ok || !(await projects.json() as Array<{ ID: string; Active: boolean }>).some(item => item.ID === project && item.Active)) throw new Error("Project unavailable");
      const response = await fetch("/api/admin/plugins?project_id=" + encodeURIComponent(project), { headers, signal });
      if (!response.ok) throw new Error("Plugin catalog unavailable");
      const catalog = await response.json() as AdminCatalog;
      const found = adminContributions(project, catalog).find(item => item.pluginID === plugin && item.version === version && item.page.id === page);
      if (!found) throw new Error("Page unavailable or permission revoked");
      if (!signal.aborted) setContribution(found);
    })().catch(reason => { if (!controller.signal.aborted) setError(String(reason)); });
    return () => controller.abort();
  }, [project, plugin, version, page]);
  const current = contribution?.projectID === project && contribution.pluginID === plugin && contribution.version === version && contribution.page.id === page ? contribution : undefined;
  return <AdminPageFrame activePage="extensions" title={current?.page.title ?? "Plugin administration"}>
    {error ? <p role="alert">{error}</p> : current ? <ActionForm key={current.href} contribution={current} /> : <p>Loading plugin page…</p>}
  </AdminPageFrame>;
}
