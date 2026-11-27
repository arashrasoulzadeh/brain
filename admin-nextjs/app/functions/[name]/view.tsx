"use client";
import { useEffect, useRef, useState } from "react";
import { useParams, usePathname } from "next/navigation";
import Link from "next/link";
import { adminUIContext, readAdminSession, projectUIURL } from "../../lib/admin-api";
import { AdminDialog, AdminProgress } from "../../components/admin-dialog";
import { AdminPageFrame } from "../../components/admin-chrome";
import { FunctionEditor } from "../function-editor";
import { FunctionExecutions } from "../function-executions";
import { byteLength, checkedFunctionDraft, checkedPublication, checkedTestResult, formatFunctionDate, functionFormError, loadFunctionDrafts, MAX_SOURCE_BYTES, normalizePermissions, parseTestInput, publicationURL, resolveFunctionScope, sendFunctionCommand, type FunctionDraft, type FunctionScope } from "../../lib/function-studio";

type Tab = "studio" | "publication" | "settings" | "logs" | "analytics";
type Operation = "save" | "publish" | "validate" | "run" | "settings";
type Notice = { kind: "success" | "error"; text: string };
type Confirmation = "save" | "publish" | "settings" | "discard" | "leave" | "reload";
const tabs: { value: Tab; label: string }[] = [{ value: "studio", label: "Code & test" }, { value: "publication", label: "Publication" }, { value: "settings", label: "Settings" }, { value: "logs", label: "Logs" }, { value: "analytics", label: "Analytics" }];

export default function FunctionDetailPage() {
  const params = useParams<{ name: string }>();
  const { namespace } = adminUIContext(usePathname());
  const name = params.name ?? "";
  return <AdminPageFrame activePage="functions" title={name || "Function"} requiredPluginID="functions"><FunctionStudio key={(namespace ?? "") + ":" + name} namespace={namespace} name={name} /></AdminPageFrame>;
}

function FunctionStudio({ name, namespace }: { name: string; namespace?: string }) {
  const [scope, setScope] = useState<FunctionScope>();
  const [draft, setDraft] = useState<FunctionDraft>();
  const [source, setSource] = useState("");
  const [version, setVersion] = useState("1");
  const [permissions, setPermissions] = useState("");
  const [publicRoute, setPublicRoute] = useState(false);
  const [tab, setTab] = useState<Tab>("studio");
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState("");
  const [reload, setReload] = useState(0);
  const [notice, setNotice] = useState<Notice>();
  const [busy, setBusy] = useState<Operation>();
  const [confirm, setConfirm] = useState<Confirmation>();
  const [testInput, setTestInput] = useState("{}");
  const [testResult, setTestResult] = useState<{ ok: boolean; text: string; duration: number }>();
  const [validation, setValidation] = useState<Notice>();
  const [logRevision, setLogRevision] = useState(0);
  const lifetime = useRef<AbortController | null>(null);
  const pending = useRef(false);
  const dirty = Boolean(draft && (source !== draft.source || version !== String(draft.version ?? 1)));
  const settingsDirty = Boolean(draft && (publicRoute !== Boolean(draft.public) || JSON.stringify(normalizePermissions(permissions)) !== JSON.stringify(normalizePermissions((draft.permissions ?? []).join(",")))));
  const hasChanges = dirty || settingsDirty;
  const formError = functionFormError(name, source, version);

  useEffect(() => {
    const controller = new AbortController();
    lifetime.current = controller;
    const token = readAdminSession();
    setLoading(true); setLoadError(""); setDraft(undefined); setScope(undefined); setNotice(undefined); setValidation(undefined); setTestResult(undefined);
    void (async () => {
      const current = await resolveFunctionScope(token, namespace, controller.signal);
      const rows = await loadFunctionDrafts(current, controller.signal);
      if (controller.signal.aborted || readAdminSession() !== token) return;
      const row = rows.find(item => item.name === name);
      setScope(current);
      if (!row) return;
      setDraft(row); setSource(row.source); setVersion(String(row.version ?? 1)); setPermissions((row.permissions ?? []).join(", ")); setPublicRoute(Boolean(row.public));
    })().catch(reason => { if (!controller.signal.aborted && readAdminSession() === token) setLoadError(String(reason)); })
      .finally(() => { if (!controller.signal.aborted && readAdminSession() === token) setLoading(false); });
    return () => { controller.abort(); if (lifetime.current === controller) lifetime.current = null; };
  }, [name, namespace, reload]);
  useEffect(() => {
    if (!hasChanges) return;
    function warn(event: BeforeUnloadEvent) { event.preventDefault(); event.returnValue = ""; }
    window.addEventListener("beforeunload", warn);
    return () => window.removeEventListener("beforeunload", warn);
  }, [hasChanges]);

  function editSource(value: string) { setSource(value); setValidation(undefined); setTestResult(undefined); }
  async function perform(operation: Operation) {
    const controller = lifetime.current;
    if (!scope || !draft || !controller || pending.current || readAdminSession() !== scope.token) return;
    if (operation !== "settings" && formError) { setNotice({ kind: "error", text: formError }); return; }
    let input: unknown;
    if (operation === "run") {
      try { input = parseTestInput(testInput); } catch (reason) { setTestResult({ ok: false, text: String(reason), duration: 0 }); return; }
    }
    pending.current = true; setBusy(operation); setNotice(undefined);
    if (operation === "run") setTestResult(undefined);
    if (operation === "validate") setValidation(undefined);
    const started = performance.now();
    const current = () => !controller.signal.aborted && readAdminSession() === scope.token;
    let saved = false;
    try {
      if (operation === "validate" || operation === "run") {
        const result = await sendFunctionCommand(scope, { operation: operation === "run" ? "functions.run" : "functions.test", name, source, ...(operation === "run" ? { input } : {}) }, controller.signal);
        const value = checkedTestResult(result, scope, name, operation === "run");
        if (!current()) return;
        if (operation === "run") setTestResult({ ok: true, text: JSON.stringify(value, null, 2) ?? "null", duration: Math.round(performance.now() - started) });
        else setValidation({ kind: "success", text: "Host preflight passed. Run a test to check actual execution." });
      } else if (operation === "settings") {
        const nextPermissions = normalizePermissions(permissions);
        await sendFunctionCommand(scope, { operation: "functions.permissions", name, permissions: nextPermissions, public: publicRoute }, controller.signal);
        if (!current()) return;
        setDraft(previous => previous ? { ...previous, permissions: nextPermissions, public: publicRoute } : previous);
        setPermissions(nextPermissions.join(", "));
        setNotice({ kind: "success", text: "Function settings saved." }); setConfirm(undefined);
      } else {
        if (operation === "save" || dirty) {
          const updated = checkedFunctionDraft(await sendFunctionCommand(scope, { operation: "functions.update", name, source, function_version: Number(version) }, controller.signal), scope, name);
          if (!current()) return;
          setDraft(updated); setSource(updated.source); setVersion(String(updated.version ?? 1)); saved = true;
        }
        if (operation === "publish") {
          const published = checkedPublication(await sendFunctionCommand(scope, { operation: "functions.publish", name }, controller.signal), scope, name);
          if (!current()) return;
          setDraft(previous => previous ? { ...previous, ...published } : previous);
          setNotice({ kind: "success", text: "Publication confirmed." });
        } else setNotice({ kind: "success", text: "Source and version saved." });
        setConfirm(undefined);
      }
    } catch (reason) {
      if (!current()) return;
      const text = reason instanceof Error ? reason.message : String(reason);
      if (operation === "run") setTestResult({ ok: false, text, duration: Math.round(performance.now() - started) });
      else if (operation === "validate") setValidation({ kind: "error", text });
      else setNotice({ kind: "error", text: (saved ? "Source saved; publication not confirmed. " : "") + text + " Refresh before retrying if the outcome is uncertain." });
    } finally {
      if (current()) { setBusy(undefined); if (operation === "run") setLogRevision(value => value + 1); }
      pending.current = false;
    }
  }

  if (loading) return <AdminProgress label="Loading function…" />;
  if (loadError) return <section className="fn-panel"><p role="alert">{loadError}</p><button className="secondary" type="button" onClick={() => setReload(value => value + 1)}>Retry</button></section>;
  if (!scope || !draft) return <section className="fn-empty"><strong>Function not found or inspection denied.</strong>{scope && <Link href={projectUIURL(scope.project, "functions")}>Back to functions</Link>}</section>;
  const back = projectUIURL(scope.project, "functions");
  const published = Boolean(draft.published_at);
  const route = window.location.origin + publicationURL(scope.namespace, name);
  const blocked = Boolean(busy);
  function openConfirmation(action: Confirmation) { setNotice(undefined); setConfirm(action); }
  function restore() { if (!draft) return; editSource(draft.source); setVersion(String(draft.version ?? 1)); setConfirm(undefined); }
  async function copy(value: string) { try { await navigator.clipboard.writeText(value); setNotice({ kind: "success", text: "Copied to clipboard." }); } catch { setNotice({ kind: "error", text: "Could not copy. Select and copy the displayed value manually." }); } }

  return <div className="fn-workspace">
    <div className="fn-breadcrumb"><Link href={back} onClick={event => { if (hasChanges || blocked) { event.preventDefault(); if (!blocked) openConfirmation("leave"); } }}>Functions</Link><span aria-hidden="true">/</span><strong>{name}</strong></div>
    <div className="fn-heading"><div><p className="fn-kicker">FUNCTION WORKSPACE</p><h2>{name}</h2><div className="fn-actions"><span className={"fn-tag " + (published ? "fn-tag-live" : "")}>{published ? "Published" : "Unpublished"}</span><span className="fn-tag">{draft.public ? "Public configured" : "Private"}</span><code>{draft.method || "—"} · v{draft.version ?? 1}</code><span role="status">{dirty ? "Unsaved code" : "Source saved"}{settingsDirty ? " · Unsaved settings" : ""}</span></div></div>
      <div className="fn-actions"><button type="button" className="secondary" disabled={blocked} onClick={() => hasChanges ? openConfirmation("reload") : setReload(value => value + 1)}>Reload saved record</button><button className="secondary" type="button" disabled={!dirty || blocked} onClick={() => openConfirmation("discard")}>Restore saved source</button><button className="secondary" type="button" disabled={!dirty || !!formError || blocked} onClick={() => published ? openConfirmation("save") : void perform("save")}>Save changes</button><button type="button" disabled={!!formError || blocked || settingsDirty} onClick={() => openConfirmation("publish")}>Publish current version</button></div></div>
    {notice && <p className={notice.kind === "error" ? "fn-error" : "notice"} role={notice.kind === "error" ? "alert" : "status"}>{notice.text}</p>}
    {busy && <AdminProgress label={{ save: "Saving source…", publish: "Publishing…", validate: "Checking source…", run: "Executing test…", settings: "Saving settings…" }[busy]} />}
    {published && <p className="fn-warning">Published functions use the current saved record. Saving source, version or access settings can change the existing endpoint immediately; publication is not an immutable deployment.</p>}
    <nav className="fn-tabs" aria-label="Function workspace views">{tabs.map(item => <button type="button" key={item.value} className={tab === item.value ? "active" : ""} aria-current={tab === item.value ? "page" : undefined} onClick={() => setTab(item.value)}>{item.label}</button>)}</nav>
    {tab === "studio" && <div className="fn-studio-grid">
      <section className="fn-panel fn-code"><div className="fn-panel-heading"><div><h3>Source code</h3><span className="muted">JavaScript module · method in source config</span></div><span className="muted">{byteLength(source).toLocaleString()} / {MAX_SOURCE_BYTES.toLocaleString()} bytes</span></div><FunctionEditor source={source} onChange={editSource} readOnly={blocked} onValidate={() => void perform("validate")} validating={busy === "validate"} validationMessage={validation?.text} validationOK={validation?.kind === "success"} /><div className="fn-toolbar"><label>Function version<input type="number" min="1" step="1" value={version} disabled={blocked} onChange={event => setVersion(event.target.value)} /></label><p className="muted">The version is metadata on the current record, not a separate release artifact.</p></div>{formError && <p role="alert" className="fn-error">{formError}</p>}</section>
      <section className="fn-panel fn-test"><h3>Test console</h3><p className="muted">Run the current unsaved source with JSON input under your operator identity. This does not publish the function. Executions are logged.</p><label>JSON input<textarea rows={8} spellCheck={false} value={testInput} disabled={blocked} onChange={event => { setTestInput(event.target.value); setTestResult(undefined); }} /></label><div className="fn-actions"><button type="button" disabled={blocked || !!formError} onClick={() => void perform("run")}>Run test</button><button type="button" className="secondary" disabled={blocked} onClick={() => { try { setTestInput(JSON.stringify(parseTestInput(testInput), null, 2)); } catch (reason) { setTestResult({ ok: false, text: String(reason), duration: 0 }); } }}>Format JSON</button></div><div className="fn-panel-heading"><h3>Result</h3>{testResult && <button className="secondary" type="button" onClick={() => void copy(testResult.text)}>Copy result</button>}</div>{testResult ? <><p role="status" className={testResult.ok ? "fn-tag fn-tag-live" : "fn-tag fn-tag-error"}>{testResult.ok ? "Execution succeeded" : "Test failed"}{testResult.duration ? " · " + testResult.duration + " ms (request)" : ""}</p><pre className="fn-output">{testResult.text}</pre></> : <div className="fn-empty"><p className="muted">{busy === "run" ? "Waiting for host execution…" : "Run a test to see the JSON result or error."}</p></div>}<p className="muted">Preflight is not execution. Return a JSON object, array or null. No arbitrary imports or network access.</p></section>
    </div>}
    {tab === "publication" && <section className="fn-panel"><h3>Current publication</h3>{published ? <><div className="fn-facts"><div><span>Version</span><strong>v{draft.version ?? 1}</strong></div><div><span>Method</span><strong>{draft.method || "—"}</strong></div><div><span>Published at</span><strong>{formatFunctionDate(draft.published_at)}</strong></div></div><label>Endpoint<input readOnly value={route} /></label><div className="fn-actions"><button className="secondary" type="button" onClick={() => void copy(route)}>Copy URL</button></div><p className="muted">Caller access is enforced by host policy. Your operator token is not a public user session.</p></> : <div className="fn-empty"><strong>Not published yet.</strong><p>Validate and run a test in Code &amp; test, then publish the current version.</p></div>}<p className="muted">The backend retains only the current publication record. Deployment history and rollback are not available here.</p></section>}
    {tab === "settings" && <section className="fn-panel"><h3>Access settings</h3><p className="muted">These stored settings do not grant host capabilities or bypass Access Management.</p><label>Permissions (comma-separated)<input value={permissions} disabled={blocked} onChange={event => setPermissions(event.target.value)} /></label><label className="checkbox-label"><input type="checkbox" checked={publicRoute} disabled={blocked} onChange={event => setPublicRoute(event.target.checked)} /> Public route enabled</label><div className="fn-actions"><button type="button" disabled={!settingsDirty || blocked} onClick={() => openConfirmation("settings")}>Save settings</button><button type="button" className="secondary" disabled={!settingsDirty || blocked} onClick={() => { setPermissions((draft.permissions ?? []).join(", ")); setPublicRoute(Boolean(draft.public)); }}>Reset settings</button></div><p className="muted">Created {formatFunctionDate(draft.created_at)} · Project {scope.project}</p></section>}
    {(tab === "logs" || tab === "analytics") && <FunctionExecutions scope={scope} name={name} revision={logRevision} analytics={tab === "analytics"} />}
    {confirm && <AdminDialog title={{ save: "Save published function", publish: "Publish current version", settings: "Update function settings", discard: "Discard unsaved source", leave: "Leave function workspace", reload: "Reload saved function" }[confirm]} scope={scope.project + " · " + name} busy={blocked} onClose={() => setConfirm(undefined)}>
      {confirm === "discard" ? <><p>Restore the saved source and version? Unsaved code changes will be lost.</p><button type="button" onClick={restore}>Discard edits</button></> : confirm === "leave" ? <><p>Unsaved code or settings will be lost.</p><Link className="button-link" href={back}>Discard and leave</Link></> : confirm === "reload" ? <><p>Reload the current host record and discard unsaved code and settings?</p><button type="button" onClick={() => { setConfirm(undefined); setReload(value => value + 1); }}>Discard and reload</button></> : <><p>{confirm === "settings" ? "Save " + (publicRoute ? "public" : "private") + " visibility and permissions: " + (normalizePermissions(permissions).join(", ") || "none") + "?" : confirm === "publish" ? "Publish version " + version + " using the source-configured HTTP method? Unsaved source and version will be saved first." : "Save source and version " + version + " to the existing published record?"}</p>{published && <p className="fn-warning">This may immediately change the live endpoint.</p>}<button type="button" disabled={blocked} onClick={() => void perform(confirm)}>{confirm === "publish" ? "Confirm publish" : confirm === "settings" ? "Confirm settings" : "Confirm save"}</button>{notice && <p role="status">{notice.text}</p>}</>}
    </AdminDialog>}
  </div>;
}
