"use client";
import { useParams } from "next/navigation";
import Link from "next/link";
import { useEffect, useRef, useState } from "react";
import { AdminPageFrame } from "../../../../../components/admin-chrome";
import { adminHeaders, readAdminSession } from "../../../../../lib/admin-api";
import { adminContributions, adminPageExtensions } from "../../../../../lib/admin-registry";
import type { AdminCatalog, AdminContribution } from "../../../../../lib/admin-registry";
import { adminPageRevision, invokeAdminAction } from "../../../../../lib/admin-sdk";

function Result({ value }: { value: unknown }) {
  if (value && typeof value === "object" && !Array.isArray(value) && "rows" in value && Array.isArray(value.rows)) {
    return <><pre>{JSON.stringify(Object.fromEntries(Object.entries(value).filter(([key]) => key !== "rows")), null, 2)}</pre><Result value={value.rows} /></>;
  }
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
  const [values, setValues] = useState<Record<string, string>>({});
  const action = contribution.page.actions.find(item => item.operation === operation);
  const fields = action?.fields ?? [];
  const [lastQuery, setLastQuery] = useState<{ operation: string; payload: unknown; revision: number }>();
  const [result, setResult] = useState<unknown>(undefined);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const pending = useRef<AbortController | null>(null);
  useEffect(() => () => pending.current?.abort(), []);
  useEffect(() => {
    if (!lastQuery || !contribution.page.refresh_on?.length) return;
    const controller = new AbortController();
    let timer: ReturnType<typeof setTimeout>;
    async function check() {
      if (controller.signal.aborted) return;
      try {
        if (!pending.current && !document.hidden) {
          const token = readAdminSession();
          const revision = await adminPageRevision(token, contribution, controller.signal);
          if (!controller.signal.aborted && !pending.current && revision !== lastQuery!.revision) {
            pending.current = controller;
            setBusy(true);
            try {
              const next = await invokeAdminAction(token, contribution, lastQuery!.operation, lastQuery!.payload, controller.signal, true);
              if (!controller.signal.aborted) {
                setResult(next);
                setLastQuery({ ...lastQuery!, revision });
              }
            } finally {
              if (pending.current === controller) pending.current = null;
              if (!controller.signal.aborted) setBusy(false);
            }
          }
        }
      } catch (reason) {
        if (!controller.signal.aborted) { setError(String(reason)); setLastQuery(undefined); }
      } finally {
        if (!controller.signal.aborted) timer = setTimeout(() => void check(), 5000);
      }
    }
    timer = setTimeout(() => void check(), 5000);
    return () => { controller.abort(); clearTimeout(timer); if (pending.current === controller) { pending.current = null; setBusy(false); } };
  }, [contribution, lastQuery]);
  async function run() {
    if (pending.current) return;
    setError(""); setResult(undefined); setLastQuery(undefined);
    let controller: AbortController | undefined;
    try {
      const parsed: unknown = fields.length ? Object.fromEntries(fields.flatMap(field => {
        const value = values[field.name] ?? "";
        if (value === "" && !field.required) return [];
        if (value === "") throw new Error(field.name + " is required");
        if (field.type === "number" && !Number.isFinite(Number(value))) throw new Error("Invalid number");
        return [[field.name, field.type === "json" ? JSON.parse(value) : field.type === "number" ? Number(value) : field.type === "boolean" ? value === "true" : value]];
      })) : JSON.parse(payload);
      const query = Boolean(action?.read_only);
      if (!query && !window.confirm("Run " + operation + " in project " + contribution.projectID + "?")) return;
      setBusy(true);
      controller = new AbortController();
      pending.current = controller;
      const token = readAdminSession();
      const revision = query && contribution.page.refresh_on?.length ? await adminPageRevision(token, contribution, controller.signal) : 0;
      const next = await invokeAdminAction(token, contribution, operation, parsed, controller.signal, query);
      if (!controller.signal.aborted) {
        setResult(next);
        if (query && contribution.page.refresh_on?.length) setLastQuery({ operation, payload: parsed, revision });
      }
    } catch (reason) { if (!controller?.signal.aborted) setError(reason instanceof Error ? reason.message : String(reason)); }
    finally { if (pending.current === controller) pending.current = null; if (!controller?.signal.aborted) setBusy(false); }
  }
  return <section className="section">
    <h2>{contribution.page.title}</h2><p>{contribution.projectID} · {contribution.pluginID}@{contribution.version}</p>
    <form onSubmit={event => { event.preventDefault(); if (!busy) void run(); }}>
      <label>Operation<select disabled={busy} value={operation} onChange={event => { setOperation(event.target.value); setResult(undefined); setError(""); setValues({}); setPayload("{}"); setLastQuery(undefined); }}>
        {contribution.page.actions.map(action => <option key={action.operation} value={action.operation}>{action.title}</option>)}
      </select></label>
      {fields.length ? fields.map(field => <label key={field.name}>{field.name}{field.required ? " *" : ""}
        {field.type === "boolean" ? <select disabled={busy} required={field.required} value={values[field.name] ?? ""} onChange={event => { setLastQuery(undefined); setValues(previous => ({ ...previous, [field.name]: event.target.value })); }}>
          <option value="">Select</option><option value="true">True</option><option value="false">False</option>
        </select> : field.type === "json" ? <textarea disabled={busy} required={field.required} rows={5} value={values[field.name] ?? ""} onChange={event => { setLastQuery(undefined); setValues(previous => ({ ...previous, [field.name]: event.target.value })); }} />
        : <input disabled={busy} required={field.required} type={field.type === "number" ? "number" : "text"} step={field.type === "number" ? "any" : undefined} value={values[field.name] ?? ""} onChange={event => { setLastQuery(undefined); setValues(previous => ({ ...previous, [field.name]: event.target.value })); }} />}
      </label>) : <label>JSON payload<textarea disabled={busy} rows={10} value={payload} onChange={event => { setLastQuery(undefined); setPayload(event.target.value); }} /></label>}
      <button disabled={busy} type="submit">{busy ? "Running…" : action?.read_only ? "Query" : "Run action"}</button>
    </form>
    {lastQuery && <p className="muted">Live refresh enabled for declared observation events.</p>}
    {error && <p role="alert">{error}</p>}
    {result !== undefined && <Result value={result} />}
  </section>;
}

export default function ExtensionPage() {
  const { project, plugin, version, page } = useParams<{ project: string; plugin: string; version: string; page: string }>();
  const [contribution, setContribution] = useState<AdminContribution>();
  const [extensions, setExtensions] = useState<AdminContribution[]>([]);
  const [error, setError] = useState("");
  useEffect(() => {
    const controller = new AbortController();
    setContribution(undefined); setExtensions([]); setError("");
    void (async () => {
      const token = readAdminSession();
      if (!token) throw new Error("Sign in to continue");
      const headers = adminHeaders(token), signal = controller.signal;
      const projects = await fetch("/api/admin/projects", { headers, signal });
      if (!projects.ok || !(await projects.json() as Array<{ ID: string; Active: boolean }>).some(item => item.ID === project && item.Active)) throw new Error("Project unavailable");
      const response = await fetch("/api/admin/plugins?project_id=" + encodeURIComponent(project), { headers, signal });
      if (!response.ok) throw new Error("Plugin catalog unavailable");
      const catalog = await response.json() as AdminCatalog;
      const contributions = adminContributions(project, catalog);
      const found = contributions.find(item => item.pluginID === plugin && item.version === version && item.page.id === page);
      if (!found) throw new Error("Page unavailable or permission revoked");
      if (!signal.aborted) { setContribution(found); setExtensions(adminPageExtensions(found, contributions)); }
    })().catch(reason => { if (!controller.signal.aborted) setError(String(reason)); });
    return () => controller.abort();
  }, [project, plugin, version, page]);
  const current = contribution?.projectID === project && contribution.pluginID === plugin && contribution.version === version && contribution.page.id === page ? contribution : undefined;
  return <AdminPageFrame activePage="extensions" title={current?.page.title ?? "Plugin administration"}>
    {error ? <p role="alert">{error}</p> : current ? <><ActionForm key={current.href} contribution={current} />{extensions.length > 0 && <section><h2>Plugin extensions</h2>{extensions.map(item => <p key={item.href}><Link href={item.href}>{item.page.title} · {item.pluginID}</Link></p>)}</section>}</> : <p>Loading plugin page…</p>}
  </AdminPageFrame>;
}
