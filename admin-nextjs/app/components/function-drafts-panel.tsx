import type { AdminCopy } from "../lib/admin-copy";
import type { FunctionDraft, SystemManifest } from "../lib/admin-types";

export function FunctionDraftsPanel({ activePage, text, functionName, setFunctionName, functionSource, setFunctionSource, functionInput, setFunctionInput, submitFunction, functionMessage, functionDrafts, system }: {
  activePage: string;
  text: AdminCopy;
  functionName: string;
  setFunctionName: (value: string) => void;
  functionSource: string;
  setFunctionSource: (value: string) => void;
  functionInput: string;
  setFunctionInput: (value: string) => void;
  submitFunction: (operation: "functions.test" | "functions.run" | "functions.create") => void;
  functionMessage: string;
  functionDrafts: FunctionDraft[];
  system: SystemManifest[];
}) {
  return (
    <>
      <section className={`section ${activePage !== "functions" ? "page-hidden" : ""}`}>
        <div className="section-heading">
          <div>
            <p className="eyebrow">{text.functions}</p>
            <h2>Custom serverless functions</h2>
            <p className="muted">
              Create project-scoped TypeScript function drafts, test source validation, run JSON
              input, and save them for host compilation.
            </p>
          </div>
        </div>
        <div className="function-editor">
          <label>
            Function name
            <input
              value={functionName}
              onChange={(event) => setFunctionName(event.target.value)}
              placeholder="hello"
            />
          </label>
          <label>
            Source code
            <textarea
              value={functionSource}
              onChange={(event) => setFunctionSource(event.target.value)}
              rows={18}
              spellCheck={false}
            />
          </label>
          <label>
            JSON input
            <textarea
              value={functionInput}
              onChange={(event) => setFunctionInput(event.target.value)}
              rows={5}
              spellCheck={false}
            />
          </label>
          <div className="playground-actions">
            <button type="button" onClick={() => void submitFunction("functions.test")}>
              Validate
            </button>
            <button type="button" onClick={() => void submitFunction("functions.run")}>
              Run test
            </button>
            <button type="button" onClick={() => void submitFunction("functions.create")}>
              Save draft
            </button>
            <small>{functionMessage}</small>
          </div>
        </div>
        <div className="table-wrap">
          <table>
            <thead>
              <tr>
                <th>Name</th>
                <th>Project</th>
                <th>Created</th>
                <th>Load</th>
              </tr>
            </thead>
            <tbody>
              {functionDrafts.map((draft) => (
                <tr key={`${draft.project_id}:${draft.name}`}>
                  <td>
                    <code>{draft.name}</code>
                  </td>
                  <td>{draft.project_id}</td>
                  <td>{new Date(draft.created_at).toLocaleString()}</td>
                  <td>
                    <button
                      type="button"
                      onClick={() => {
                        setFunctionName(draft.name);
                        setFunctionSource(draft.source);
                      }}
                    >
                      Load
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        <div className="plugin-grid">
          {system
            .filter((manifest) => manifest.ID === "functions")
            .map((manifest) => (
              <article className="plugin" key={`functions-${manifest.ID}`}>
                <h3>{manifest.Title}</h3>
                <p className="muted">{manifest.Description}</p>
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
