"use client";
import { useState } from "react";
import { routeIdentity, searchRoutes } from "../lib/admin-playground";
import type { ProjectHTTPRoute } from "../lib/admin-types";
import type { AdminCopy } from "../lib/admin-copy";

export function RoutesPanel({ activePage, text, playgroundRoute, setPlaygroundRoute, playgroundPayload, setPlaygroundPayload, testPlayground, sendPlayground, playgroundMessage, selectedProjectID, projectRoutes }: {
  activePage: string;
  text: AdminCopy;
  playgroundRoute: string;
  setPlaygroundRoute: (value: string) => void;
  playgroundPayload: string;
  setPlaygroundPayload: (value: string) => void;
  testPlayground: () => void;
  sendPlayground: () => void;
  playgroundMessage: string;
  selectedProjectID: string;
  projectRoutes: ProjectHTTPRoute[];
}) {
  const [query, setQuery] = useState("");
  const [method, setMethod] = useState("");
  const [auth, setAuth] = useState("");
  const filtered = searchRoutes(projectRoutes, query, method, auth);
  const selected = projectRoutes.find(route => routeIdentity(route) === playgroundRoute);
  return (
    <>
      <section className={`section ${activePage !== "routes" ? "page-hidden" : ""}`}>
        <div className="section-heading">
          <div>
            <p className="eyebrow">{text.playground}</p>
            <h2>{text.playground}</h2>
          </div>
        </div>
        <div className="playground">
          <label>
            {text.selectRoute}
            <select
              value={playgroundRoute}
              onChange={(event) => setPlaygroundRoute(event.target.value)}
            >
              <option value="">—</option>
              {projectRoutes.map((route) => (
                <option key={`${route.plugin_id}:${route.version}:${route.operation}:${route.method}:${route.path}`} value={routeIdentity(route)}>
                  {route.method} {route.path} · {route.operation}
                </option>
              ))}
            </select>
          </label>
          <label>
            {text.payload}
            <textarea
              aria-label={text.payload}
              value={playgroundPayload}
              onChange={(event) => setPlaygroundPayload(event.target.value)}
              rows={8}
            />
          </label>
          {selected && <p className="muted">{selected.method} {selected.path} · {selected.auth} · {selected.capability}</p>}
          <div className="playground-actions">
            <button type="button" onClick={testPlayground}>
              {text.test}
            </button>
            <button type="button" className="secondary" onClick={() => { try { setPlaygroundPayload(JSON.stringify(JSON.parse(playgroundPayload || "{}"), null, 2)); } catch { testPlayground(); } }}>Format JSON</button>
            <button type="button" className="secondary" onClick={() => setPlaygroundPayload("{}")}>Reset payload</button>
            <button type="button" disabled={!selected || !selectedProjectID} onClick={sendPlayground}>
              {text.send}
            </button>
            <small role="status">{playgroundMessage}</small>
          </div>
          <pre>
            {playgroundRoute
              ? JSON.stringify(
                  {
                    project_id: selectedProjectID,
                    plugin_id: playgroundRoute.split(":")[0],
                    version: playgroundRoute.split(":")[1],
                    operation: playgroundRoute.split(":")[2],
                    payload: (() => { try { return JSON.parse(playgroundPayload || "{}"); } catch { return playgroundPayload; } })(),
                  },
                  null,
                  2,
                )
              : text.selectRoute}
          </pre>
        </div>
        <div className="section-heading">
          <div>
            <p className="eyebrow">{text.routesPage}</p>
            <h2>{text.routesPage}</h2>
          </div>
          <span className="count">
            {filtered.length} / {projectRoutes.length}
          </span>
        </div>
        <div className="lifecycle-actions"><label>Search routes<input type="search" value={query} onChange={event => setQuery(event.target.value)} placeholder="Path, plugin or operation" /></label><label>HTTP method<select value={method} onChange={event => setMethod(event.target.value)}><option value="">All methods</option>{[...new Set(projectRoutes.map(route => route.method))].sort().map(value => <option key={value}>{value}</option>)}</select></label><label>Authentication<select value={auth} onChange={event => setAuth(event.target.value)}><option value="">All modes</option><option value="public">Public</option><option value="user">User</option><option value="operator">Operator</option></select></label></div>
        <div className="plugin-grid">
          {filtered.map((route) => (
            <article className="plugin" key={`${route.plugin_id}@${route.version}:${route.method}:${route.path}`}>
              <h3>{route.method} {route.path}</h3>
              <div className="route-list">
                <code>{route.plugin_id}@{route.version} · {route.operation}</code>
                <small>{route.capability} · {route.auth} · {route.function}</small>
                <button type="button" className="secondary" onClick={() => { setPlaygroundRoute(routeIdentity(route)); setPlaygroundPayload("{}"); }}>Use in playground</button>
              </div>
            </article>
          ))}
          {!!projectRoutes.length && !filtered.length && <p className="muted">No matching routes.</p>}
          {selectedProjectID && projectRoutes.length === 0 && <p className="muted">No enabled plugin HTTP routes for this project.</p>}
        </div>
      </section>
    </>
  );
}
