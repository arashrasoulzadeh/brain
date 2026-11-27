"use client";
import { adminHeaders, readAdminSession, preferredAdminProject, projectAdminURL, systemAdminURL, projectUIURL, adminUIContext } from "../../lib/admin-api";

import { FormEvent, useState } from "react";
import { usePathname, useRouter } from "next/navigation";
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

  async function save(event: FormEvent) {
    event.preventDefault();
    const token = readAdminSession();
    const projects = await (await fetch(systemAdminURL("projects"), { headers: adminHeaders(token) })).json() as { ID: string; Slug?: string; Active: boolean }[];
    const project = preferredAdminProject(projects, namespace)?.ID ?? "";
    if (!project) { setMessage("Project unavailable"); return; }
    const response = await fetch(projectAdminURL(project, "functions"), { method: "POST", headers: adminHeaders(token, { "content-type": "application/json" }), body: JSON.stringify({ operation: "functions.create", project_id: project, name, source, function_version: Number(version) || 1, public: publicRoute, permissions: permissions.split(",").map((value) => value.trim()).filter(Boolean) }) });
    if (!response.ok) { setMessage(await response.text()); return; }
    router.push(projectUIURL(project, "functions", name));
  }

  return <AdminPageFrame activePage="functions" title="Create function" requiredPluginID="functions"><form className="function-studio" onSubmit={save}><label>Function name<input value={name} onChange={(event) => setName(event.target.value)} required placeholder="hello" /></label><label>Source code<FunctionEditor source={source} onChange={setSource} /></label><div className="function-settings"><label>Version<input type="number" min="1" value={version} onChange={(event) => setVersion(event.target.value)} /></label><label>Permissions<input value={permissions} onChange={(event) => setPermissions(event.target.value)} /></label><label className="checkbox-label"><input type="checkbox" checked={publicRoute} onChange={(event) => setPublicRoute(event.target.checked)} /> Public call route</label></div><p className="muted">Set the HTTP method in the source config, for example <code>export const config = &#123; method: &apos;POST&apos; &#125;</code>.</p><div className="playground-actions"><button type="submit">Create function</button><small>{message}</small></div></form></AdminPageFrame>;
}
