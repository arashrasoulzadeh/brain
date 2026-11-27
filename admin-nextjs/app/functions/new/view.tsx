"use client";
import { adminHeaders, readAdminSession, preferredAdminProject, projectAdminURL, systemAdminURL, projectUIURL, adminUIContext } from "../../lib/admin-api";

import { FormEvent, useRef, useState } from "react";
import { usePathname, useRouter } from "next/navigation";
import { AdminProgress } from "../../components/admin-dialog";
import { AdminPageFrame } from "../../components/admin-chrome";
import { FunctionEditor } from "../function-editor";

export default function NewFunctionPage() {
  const router = useRouter();
  const { namespace } = adminUIContext(usePathname());
  const [name, setName] = useState("");
  const [source, setSource] = useState("export const config = { method: 'POST' };\n\nexport default async function handler(input) {\n  return input;\n}\n");
  const [version, setVersion] = useState("1");
  const [publicRoute, setPublicRoute] = useState(false);
  const [permissions, setPermissions] = useState("functions.invoke");
  const [message, setMessage] = useState("");
  const [busy, setBusy] = useState(false);
  const pending = useRef(false);

  async function save(event: FormEvent) {
    event.preventDefault();
    if (pending.current) return;
    pending.current = true; setBusy(true); setMessage("");
    try {
      const token = readAdminSession();
      if (!token || !namespace) throw new Error("Select an active project and sign in");
      const projectResponse = await fetch(systemAdminURL("projects"), { headers: adminHeaders(token) });
      if (!projectResponse.ok) throw new Error("Could not load projects");
      const projects = (await projectResponse.json() as { ID: string; Slug?: string; Active: boolean }[] | null) ?? [];
      const project = preferredAdminProject(projects, namespace)?.ID;
      if (!project) throw new Error("Project unavailable");
      const number = Number(version);
      if (!Number.isSafeInteger(number) || number < 1) throw new Error("Version must be a positive integer");
      const response = await fetch(projectAdminURL(project, "functions"), { method: "POST", headers: adminHeaders(token, { "content-type": "application/json" }), body: JSON.stringify({ operation: "functions.create", project_id: project, name: name.trim(), source, function_version: number, public: publicRoute, permissions: [...new Set(permissions.split(",").map(value => value.trim()).filter(Boolean))] }) });
      if (!response.ok) throw new Error(await response.text());
      router.push(projectUIURL(project, "functions", name.trim()));
    } catch (reason) { setMessage(String(reason)); }
    finally { pending.current = false; setBusy(false); }
  }

  return <AdminPageFrame activePage="functions" title="Create function" requiredPluginID="functions"><form className="function-studio" onSubmit={save}><label>Function name<input disabled={busy} value={name} onChange={(event) => setName(event.target.value)} required placeholder="hello" /></label><label>Source code<FunctionEditor source={source} onChange={setSource} readOnly={busy} /></label><div className="function-settings"><label>Version<input disabled={busy} type="number" min="1" step="1" value={version} onChange={(event) => setVersion(event.target.value)} /></label><label>Permissions<input disabled={busy} value={permissions} onChange={(event) => setPermissions(event.target.value)} /></label><label className="checkbox-label"><input disabled={busy} type="checkbox" checked={publicRoute} onChange={(event) => setPublicRoute(event.target.checked)} /> Public call route</label></div><p className="muted">Set the HTTP method in the source config, for example <code>export const config = &#123; method: &apos;POST&apos; &#125;</code>.</p><div className="playground-actions"><button type="submit" disabled={busy}>Create function</button>{message && <small role="alert">{message}</small>}{busy && <AdminProgress label="Creating function…" />}</div></form></AdminPageFrame>;
}
