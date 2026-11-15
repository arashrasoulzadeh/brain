"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { AdminPageFrame } from "../components/admin-chrome";

type Project = { ID: string; Active: boolean };

export default function WorkflowsPage() {
  const router = useRouter();
  const [session, setSession] = useState("");
  const [project, setProject] = useState("");
  const [operation, setOperation] = useState("functions.invoke");
  const [payload, setPayload] = useState("{}");
  const [result, setResult] = useState("");
  const [running, setRunning] = useState(false);

  useEffect(() => {
    const token = window.sessionStorage.getItem("brain-admin-session") ?? "";
    if (!token) {
      router.replace("/");
      return;
    }
    setSession(token);
    void fetch("/api/admin/projects", { headers: { authorization: `Bearer ${token}` } }).then(async (response) => {
      if (!response.ok) throw new Error("Could not load projects.");
      const projects = await response.json() as Project[];
      setProject(projects.find((candidate) => candidate.Active)?.ID ?? "");
    }).catch(() => {
      window.sessionStorage.removeItem("brain-admin-session");
      router.replace("/");
    });
  }, [router]);

  async function runWorkflow() {
    if (!project) {
      setResult("Select an active project first.");
      return;
    }
    setRunning(true);
    try {
      const parsedPayload = JSON.parse(payload || "{}");
      const response = await fetch("/api/admin/actions", { method: "POST", headers: { authorization: `Bearer ${session}`, "content-type": "application/json" }, body: JSON.stringify({ operation: "plugin.invoke", project_id: project, plugin_id: "functions", version: "0.1.0", target_operation: operation, payload: parsedPayload }) });
      const body = await response.text();
      setResult(body || (response.ok ? "Operation completed." : `Request failed (${response.status}).`));
    } catch (error) {
      setResult(error instanceof Error ? error.message : "Invalid workflow payload.");
    } finally {
      setRunning(false);
    }
  }

  return <AdminPageFrame activePage="workflows" title="Workflows" projectID={project} onSignOut={() => { window.sessionStorage.removeItem("brain-admin-session"); router.replace("/"); }}><p className="muted">Run project-scoped Functions plugin operations through the host.</p><div className="mt-6 grid gap-5 rounded-2xl border border-slate-200 bg-slate-50 p-5 md:grid-cols-2"><label className="m-0 flex flex-col gap-2 text-sm font-semibold text-slate-700">Operation<select value={operation} onChange={(event) => setOperation(event.target.value)}><option value="functions.invoke">functions.invoke</option><option value="functions.publish">functions.publish</option></select></label><label className="m-0 flex flex-col gap-2 text-sm font-semibold text-slate-700">JSON payload<textarea value={payload} onChange={(event) => setPayload(event.target.value)} rows={10} /></label><div className="md:col-span-2 flex items-center gap-3"><button type="button" disabled={running} onClick={() => void runWorkflow()}>{running ? "Running…" : "Run workflow"}</button>{result && <pre className="m-0 max-w-full overflow-auto rounded-lg bg-slate-900 p-3 text-xs text-slate-100">{result}</pre>}</div></div></AdminPageFrame>;
}
