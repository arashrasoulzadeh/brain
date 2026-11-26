import type { AdminCopy } from "../lib/admin-copy";
import type { SystemManifest } from "../lib/admin-types";

export function RoutesPanel({ activePage, text, playgroundRoute, setPlaygroundRoute, system, playgroundPayload, setPlaygroundPayload, testPlayground, sendPlayground, playgroundMessage, selectedProjectID }: {
  activePage: string;
  text: AdminCopy;
  playgroundRoute: string;
  setPlaygroundRoute: (value: string) => void;
  system: SystemManifest[];
  playgroundPayload: string;
  setPlaygroundPayload: (value: string) => void;
  testPlayground: () => void;
  sendPlayground: () => void;
  playgroundMessage: string;
  selectedProjectID: string;
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
              {system.flatMap(
                (manifest) =>
                  manifest.Routes?.map((route) => (
                    <option
                      key={`${manifest.ID}:${route.Operation}`}
                      value={`${manifest.ID}:${route.Operation}`}
                    >
                      {manifest.ID} · {route.Operation}
                    </option>
                  )) ?? [],
              )}
            </select>
          </label>
          <label>
            {text.payload}
            <textarea
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
            <small>{playgroundMessage}</small>
          </div>
          <pre>
            {playgroundRoute
              ? JSON.stringify(
                  {
                    project_id: selectedProjectID,
                    operation: playgroundRoute.split(":")[1],
                    payload: playgroundPayload,
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
            {system.reduce((total, manifest) => total + (manifest.Routes?.length ?? 0), 0)}
          </span>
        </div>
        <div className="plugin-grid">
          {system.map((manifest) => (
            <article className="plugin" key={`routes-${manifest.ID}`}>
              <h3>{manifest.Title}</h3>
              <div className="route-list">
                {manifest.Routes?.map((route) => (
                  <div key={route.Operation}>
                    <code>{route.Operation}</code>
                    <small>{route.Capability}</small>
                  </div>
                ))}
              </div>
            </article>
          ))}
        </div>
      </section>
    </>
  );
}
