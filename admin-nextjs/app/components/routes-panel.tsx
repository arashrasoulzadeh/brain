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
                <option key={`${route.plugin_id}:${route.operation}:${route.method}:${route.path}`} value={`${route.plugin_id}:${route.version}:${route.operation}`}>
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
          <div className="playground-actions">
            <button type="button" onClick={testPlayground}>
              {text.test}
            </button>
            <button type="button" onClick={sendPlayground}>
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
            {projectRoutes.length}
          </span>
        </div>
        <div className="plugin-grid">
          {projectRoutes.map((route) => (
            <article className="plugin" key={`${route.plugin_id}@${route.version}:${route.method}:${route.path}`}>
              <h3>{route.method} {route.path}</h3>
              <div className="route-list">
                <code>{route.plugin_id}@{route.version} · {route.operation}</code>
                <small>{route.capability} · {route.auth} · {route.function}</small>
              </div>
            </article>
          ))}
          {selectedProjectID && projectRoutes.length === 0 && <p className="muted">No enabled plugin HTTP routes for this project.</p>}
        </div>
      </section>
    </>
  );
}
