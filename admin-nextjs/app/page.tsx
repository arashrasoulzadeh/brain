"use client";
import { FormEvent, useState } from "react";
type Project = { ID: string; Active: boolean };
type Language = "en" | "fa";
const copy = {
  en: {
    eyebrow: "BRAIN CONTROL PLANE",
    welcome: "Welcome back",
    subtitle: "Securely manage projects and plugins.",
    username: "Operator username",
    secret: "Secret",
    signIn: "Sign in",
    projects: "Projects",
    signOut: "Sign out",
    newProject: "New project ID",
    create: "Create project",
    empty: "No projects yet.",
    active: "Active",
    inactive: "Inactive",
    invalid: "Invalid operator credentials",
  },
  fa: {
    eyebrow: "پنل مدیریت برین",
    welcome: "خوش آمدید",
    subtitle: "پروژه‌ها و افزونه‌ها را با امنیت مدیریت کنید.",
    username: "نام کاربری مدیر",
    secret: "رمز عبور",
    signIn: "ورود",
    projects: "پروژه‌ها",
    signOut: "خروج",
    newProject: "شناسه پروژه جدید",
    create: "ساخت پروژه",
    empty: "هنوز پروژه‌ای وجود ندارد.",
    active: "فعال",
    inactive: "غیرفعال",
    invalid: "اطلاعات ورود نادرست است",
  },
};
export default function AdminPage() {
  const [language, setLanguage] = useState<Language>("en");
  const [session, setSession] = useState("");
  const [projects, setProjects] = useState<Project[]>([]);
  const [username, setUsername] = useState("");
  const [secret, setSecret] = useState("");
  const [projectID, setProjectID] = useState("");
  const [error, setError] = useState("");
  const text = copy[language];
  async function loadProjects(token: string) {
    const response = await fetch("/api/admin/projects", {
      headers: { authorization: `Bearer ${token}` },
    });
    if (!response.ok) throw new Error("Could not load projects");
    setProjects((await response.json()) as Project[]);
  }
  async function signIn(event: FormEvent) {
    event.preventDefault();
    setError("");
    try {
      const response = await fetch("/api/admin/session", {
        method: "POST",
        headers: { "content-type": "application/json" },
        body: JSON.stringify({ username, secret }),
      });
      if (!response.ok) throw new Error(text.invalid);
      const token = ((await response.json()) as { session: string }).session;
      setSession(token);
      await loadProjects(token);
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Sign in failed");
    }
  }
  async function createProject(event: FormEvent) {
    event.preventDefault();
    setError("");
    try {
      const response = await fetch("/api/admin/actions", {
        method: "POST",
        headers: { authorization: `Bearer ${session}`, "content-type": "application/json" },
        body: JSON.stringify({ operation: "project.create", project_id: projectID }),
      });
      if (!response.ok) throw new Error(await response.text());
      setProjectID("");
      await loadProjects(session);
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Could not create project");
    }
  }
  const switcher = (
    <button
      className="language"
      type="button"
      onClick={() => setLanguage(language === "en" ? "fa" : "en")}
    >
      {language === "en" ? "فارسی" : "English"}
    </button>
  );
  if (!session)
    return (
      <main className="shell" dir={language === "fa" ? "rtl" : "ltr"}>
        <section className="card login">
          {switcher}
          <p className="eyebrow">{text.eyebrow}</p>
          <h1>{text.welcome}</h1>
          <p className="muted">{text.subtitle}</p>
          <form onSubmit={signIn}>
            <label>
              {text.username}
              <input
                value={username}
                onChange={(event) => setUsername(event.target.value)}
                required
              />
            </label>
            <label>
              {text.secret}
              <input
                type="password"
                value={secret}
                onChange={(event) => setSecret(event.target.value)}
                required
              />
            </label>
            <button>{text.signIn}</button>
            {error && <p className="error">{error}</p>}
          </form>
        </section>
      </main>
    );
  return (
    <main className="shell" dir={language === "fa" ? "rtl" : "ltr"}>
      <section className="card">
        <div className="top">
          <div>
            <p className="eyebrow">{text.eyebrow}</p>
            <h1>{text.projects}</h1>
          </div>
          <div>
            {switcher}
            <button className="secondary" onClick={() => setSession("")}>
              {text.signOut}
            </button>
          </div>
        </div>
        <form className="create" onSubmit={createProject}>
          <input
            placeholder={text.newProject}
            value={projectID}
            onChange={(event) => setProjectID(event.target.value)}
            required
          />
          <button>{text.create}</button>
        </form>
        {error && <p className="error">{error}</p>}
        <div>
          {projects.length ? (
            projects.map((project) => (
              <div className="project" key={project.ID}>
                <span>{project.ID}</span>
                <small>{project.Active ? text.active : text.inactive}</small>
              </div>
            ))
          ) : (
            <p className="muted">{text.empty}</p>
          )}
        </div>
      </section>
    </main>
  );
}
