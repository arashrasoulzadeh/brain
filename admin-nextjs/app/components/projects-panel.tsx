import type { FormEvent } from "react";
import type { AdminCopy } from "../lib/admin-copy";
import type { Project } from "../lib/admin-types";

export function ProjectsPanel({ activePage, text, selectedProjectID, projects, projectID, setProjectID, createProject, setProjectActive, uninstallProject, error, onSelectProject }: {
  activePage: string;
  text: AdminCopy;
  selectedProjectID: string;
  projects: Project[];
  projectID: string;
  setProjectID: (value: string) => void;
  createProject: (event: FormEvent) => void;
  setProjectActive: (project: Project, active: boolean) => void;
  uninstallProject: (project: Project) => void;
  error: string;
  onSelectProject: (id: string) => void;
}) {
  return (
    <>
      <div
        className={activePage === "projects" ? "project-selector" : "project-selector page-hidden"}
      >
        <label htmlFor="selected-project">{text.project}</label>
        <select
          id="selected-project"
          value={selectedProjectID}
          onChange={(event) => onSelectProject(event.target.value)}
        >
          <option value="">—</option>
          {projects.map((project) => (
            <option key={project.ID} value={project.ID} disabled={!project.Active}>
              {project.ID}
              {project.Active ? "" : " (inactive)"}
            </option>
          ))}
        </select>
      </div>
      <form
        className={activePage === "projects" ? "create" : "create page-hidden"}
        onSubmit={createProject}
      >
        <input
          placeholder={text.newProject}
          value={projectID}
          onChange={(event) => setProjectID(event.target.value)}
          required
        />
        <button>{text.create}</button>
      </form>
      {error && <p className="error">{error}</p>}
      <div className={activePage === "projects" ? "" : "page-hidden"}>
        {projects.length ? (
          <div className="project-table">
            <div className="project-table-head">
              <span>{text.project}</span>
              <span>{text.active}</span>
              <span>{text.navigation}</span>
            </div>
            {projects.map((project) => (
              <div className="project-table-row" key={project.ID}>
                <strong>{project.ID}</strong>
                <span className={project.Active ? "success" : "failure"}>
                  {project.Active ? text.active : text.inactive}
                </span>
                <span className="project-actions">
                  {project.Active ? (
                    <button
                      className="secondary"
                      onClick={() => void setProjectActive(project, false)}
                    >
                      {text.deactivate}
                    </button>
                  ) : (
                    <button onClick={() => void setProjectActive(project, true)}>
                      {text.activate}
                    </button>
                  )}
                  {project.Active && (
                    <button className="danger" onClick={() => uninstallProject(project)}>
                      {text.uninstall}
                    </button>
                  )}
                </span>
              </div>
            ))}
          </div>
        ) : (
          <p className="muted">{text.empty}</p>
        )}
      </div>
    </>
  );
}
