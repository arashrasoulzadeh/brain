"use client";
import { useEffect, useRef, useState } from "react";
import Link from "next/link";
import { readAdminSession } from "../lib/admin-api";
import type { AdminContribution } from "../lib/admin-registry";
import { adminEmbeddedExtensions } from "../lib/admin-registry";
import { adminPageRevision, invokeAdminAction } from "../lib/admin-sdk";
import { AdminResultRenderer } from "./admin-results";
import { AdminInventory } from "./admin-inventory";
import { ModelPanel } from "./model-panel";
import { FileUploadField, FileDownloadResult } from "./admin-file-controls";

export function ActionForm({ contribution }: { contribution: AdminContribution }) {
  const [operation, setOperation] = useState(contribution.page.actions[0].operation);
  const [payload, setPayload] = useState("{}");
  const [values, setValues] = useState<Record<string, string>>({});
  const action = contribution.page.actions.find(item => item.operation === operation);
  const fields = action?.fields ?? [];
  const [lastQuery, setLastQuery] = useState<{ operation: string; payload: unknown; revision: number }>();
  const [result, setResult] = useState<unknown>(undefined);
  const [inventoryRefresh, setInventoryRefresh] = useState(0);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const [fileReading, setFileReading] = useState(false);
  const filesPage = contribution.pluginID === "files" && contribution.page.id === "operations";
  const uploadField = filesPage && operation === "files.upload";
  const controlsDisabled = busy || fileReading;
  function updateValues(update: (previous: Record<string, string>) => Record<string, string>) {
    setLastQuery(undefined); setResult(undefined); setError(""); setValues(update);
  }
  function selectOperation(next: string) {
    if (controlsDisabled || next === operation) return;
    setOperation(next); setResult(undefined); setError(""); setLastQuery(undefined); setPayload("{}");
    setValues(previous => filesPage && next !== "files.upload" && previous.key ? { key: previous.key } : {});
  }
  const pending = useRef<AbortController | null>(null);
  useEffect(() => () => pending.current?.abort(), []);
  useEffect(() => {
    if (!lastQuery) return;
    const controller = new AbortController();
    let timer: ReturnType<typeof setTimeout>;
    async function check() {
      if (controller.signal.aborted) return;
      try {
        if (!pending.current && !document.hidden) {
          const token = readAdminSession();
          const revision = await adminPageRevision(token, contribution, controller.signal, lastQuery);
          if (!controller.signal.aborted && !pending.current && contribution.page.refresh_on?.length && revision !== lastQuery!.revision) {
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
        if (!controller.signal.aborted) { setResult(undefined); setError(String(reason)); setLastQuery(undefined); }
      } finally {
        if (!controller.signal.aborted) timer = setTimeout(() => void check(), 5000);
      }
    }
    timer = setTimeout(() => void check(), 5000);
    return () => { controller.abort(); clearTimeout(timer); if (pending.current === controller) { pending.current = null; setBusy(false); } };
  }, [contribution, lastQuery]);
  async function run() {
    if (pending.current || fileReading) return;
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
      const confirmation = uploadField ? "Upload file " + (values.key ?? "") + " to project " + contribution.projectID + "? An existing file with the same key will be replaced and unpublished." : "Run " + operation + " in project " + contribution.projectID + "?";
      if (!query && !window.confirm(confirmation)) return;
      setBusy(true);
      controller = new AbortController();
      pending.current = controller;
      const token = readAdminSession();
      const revision = query && contribution.page.refresh_on?.length ? await adminPageRevision(token, contribution, controller.signal, { operation, payload: parsed }) : 0;
      const next = await invokeAdminAction(token, contribution, operation, parsed, controller.signal, query);
      if (!controller.signal.aborted) {
        setResult(next);
        if (!query) setInventoryRefresh(value => value + 1);
        if (query) setLastQuery({ operation, payload: parsed, revision });
      }
    } catch (reason) { if (!controller?.signal.aborted) setError(reason instanceof Error ? reason.message : String(reason)); }
    finally { if (pending.current === controller) pending.current = null; if (!controller?.signal.aborted) setBusy(false); }
  }
  return <section className="section admin-action-panel">
    <div className="panel-heading"><h2>{contribution.page.title}</h2><p className="muted">{contribution.projectID} · {contribution.pluginID}@{contribution.version}</p></div>
    {filesPage && <div className="lifecycle-actions" aria-label="File actions">{contribution.page.actions.map(item => <button key={item.operation} type="button" className="secondary" disabled={controlsDisabled} aria-pressed={operation === item.operation} onClick={() => selectOperation(item.operation)}>{item.title}</button>)}</div>}
    {contribution.page.inventory && <AdminInventory contribution={contribution} refresh={inventoryRefresh} disabled={controlsDisabled}
      onSelect={fields.some(field => field.name === contribution.page.inventory?.select_field && field.type === "string") ? key => {
        setLastQuery(undefined); setResult(undefined); setError("");
        setValues(previous => ({ ...previous, [contribution.page.inventory!.select_field!]: key }));
      } : undefined} />}
    <form className="admin-operation-form" onSubmit={event => { event.preventDefault(); if (!controlsDisabled) void run(); }}>
      <label>Operation<select disabled={controlsDisabled} value={operation} onChange={event => selectOperation(event.target.value)}>
        {contribution.page.actions.map(action => <option key={action.operation} value={action.operation}>{action.title}</option>)}
      </select></label>
      {fields.length ? fields.map(field => uploadField && field.name === "data" && field.type === "string" ? <FileUploadField key={field.name} value={values.data ?? ""} disabled={controlsDisabled} onReading={setFileReading}
        onData={data => updateValues(previous => ({ ...previous, data }))}
        onFile={(name, data) => updateValues(previous => ({ ...previous, key: previous.key || name, data }))} /> : <label className={field.type === "json" ? "wide-field" : undefined} key={field.name}>{field.name}{field.required ? " *" : ""}
        {field.type === "boolean" ? <select disabled={controlsDisabled} required={field.required} value={values[field.name] ?? ""} onChange={event => { setLastQuery(undefined); setResult(undefined); setError(""); setValues(previous => ({ ...previous, [field.name]: event.target.value })); }}>
          <option value="">Select</option><option value="true">True</option><option value="false">False</option>
        </select> : field.type === "json" ? <textarea disabled={controlsDisabled} required={field.required} rows={5} value={values[field.name] ?? ""} onChange={event => { setLastQuery(undefined); setResult(undefined); setError(""); setValues(previous => ({ ...previous, [field.name]: event.target.value })); }} />
        : <input disabled={controlsDisabled} required={field.required} type={field.type === "number" ? "number" : "text"} step={field.type === "number" ? "any" : undefined} value={values[field.name] ?? ""} onChange={event => { setLastQuery(undefined); setResult(undefined); setError(""); setValues(previous => ({ ...previous, [field.name]: event.target.value })); }} />}
      </label>) : action?.read_only ? null : <label className="wide-field">JSON payload<textarea disabled={controlsDisabled} rows={10} value={payload} onChange={event => { setLastQuery(undefined); setResult(undefined); setError(""); setPayload(event.target.value); }} /></label>}
      <div className="operation-submit"><button disabled={controlsDisabled} type="submit">{fileReading ? "Reading file…" : busy ? "Running…" : action?.read_only ? "Query" : "Run action"}</button>{lastQuery && <span className="muted live-indicator">{contribution.page.refresh_on?.length ? "Auto-refresh active" : "Access monitored"}</span>}</div>
    </form>
    {error && <p role="alert">{error}</p>}
    {result !== undefined && <div className="admin-result-panel">{filesPage && operation === "files.download" ? <FileDownloadResult result={result} /> : <AdminResultRenderer key={operation} value={result} spec={action?.result} />}</div>}
  </section>;
}

export function AdminPageRenderer({ contribution, extensions = [] }: { contribution: AdminContribution; extensions?: AdminContribution[] }) {
  const embedded = adminEmbeddedExtensions(contribution, extensions);
  const tabs = [contribution, ...embedded];
  const [selected, setSelected] = useState(contribution.href);
  const active = tabs.find(item => item.href === selected) ?? contribution;
  const links = extensions.filter(item => !embedded.some(tab => tab.href === item.href));
  return <>{embedded.length > 0 && <div role="tablist" aria-label="Plugin pages" className="lifecycle-actions">{tabs.map(item => <button key={item.href} type="button" role="tab" aria-selected={active.href === item.href} onClick={() => setSelected(item.href)}>{item.page.title} · {item.pluginID}</button>)}</div>}
    <div role={embedded.length > 0 ? "tabpanel" : undefined}>{active.page.model ? <ModelPanel key={active.href} contribution={active} /> : <ActionForm key={active.href} contribution={active} />}</div>
    {links.length > 0 && <section><h2>Plugin extensions</h2>{links.map(item => <p key={item.href}><Link href={item.href}>{item.page.title} · {item.pluginID}</Link></p>)}</section>}</>;
}
