"use client";

import { FormEvent, useState } from "react";
import { useRouter } from "next/navigation";
import { AdminPageFrame } from "../../components/admin-chrome";
import { FunctionEditor } from "../function-editor";

export default function NewFunctionPage() {
  const router = useRouter();
  const [name, setName] = useState("");
  const [source, setSource] = useState("export default async function handler(input) {\n  return input;\n}\n");
  const [version, setVersion] = useState("1");
  const [publicRoute, setPublicRoute] = useState(false);
  const [permissions, setPermissions] = useState("functions.invoke");
  const [method, setMethod] = useState("POST");
  const [message, setMessage] = useState("");

  async function save(event: FormEvent) {
    event.preventDefault();
    const token = window.sessionStorage.getItem("brain-admin-session") ?? "";
    const projects = await (await fetch("/api/admin/projects", { headers: { authorization: `Bearer ${token}` } })).json() as { ID: string; Active: boolean }[];
    const project = projects.find((candidate) => candidate.Active)?.ID ?? "";
    const response = await fetch("/api/admin/functions", { method: "POST", headers: { authorization: `Bearer ${token}`, "content-type": "application/json" }, body: JSON.stringify({ operation: "functions.create", project_id: project, name, source, function_version: Number(version) || 1, method, public: publicRoute, permissions: permissions.split(",").map((value) => value.trim()).filter(Boolean) }) });
    if (!response.ok) { setMessage(await response.text()); return; }
    router.push(`/functions/${encodeURIComponent(name)}`);
  }

  return <AdminPageFrame activePage="functions" title="Create function"><form className="function-studio" onSubmit={save}><label>Function name<input value={name} onChange={(event) => setName(event.target.value)} required placeholder="hello" /></label><label>Source code<FunctionEditor source={source} onChange={setSource} /></label><div className="function-settings"><label>Version<input type="number" min="1" value={version} onChange={(event) => setVersion(event.target.value)} /></label><label>Method<select value={method} onChange={(event) => setMethod(event.target.value)}><option>GET</option><option>POST</option><option>PUT</option><option>PATCH</option></select></label><label>Permissions<input value={permissions} onChange={(event) => setPermissions(event.target.value)} /></label><label className="checkbox-label"><input type="checkbox" checked={publicRoute} onChange={(event) => setPublicRoute(event.target.checked)} /> Public call route</label></div><div className="playground-actions"><button type="submit">Create function</button><small>{message}</small></div></form></AdminPageFrame>;
}
