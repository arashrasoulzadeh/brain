"use client";

import Link from "next/link";

type AdminSidebarProps = { activePage: string; projectID?: string; onSignOut?: () => void; title?: string; visiblePages?: string[] };

export function AdminSidebar({ activePage, projectID, onSignOut, title = "Projects", visiblePages }: AdminSidebarProps) {
  const links = [["projects", "Projects"], ["plugins", "Plugins"], ["system", "System capabilities"], ["audit", "Audit log"], ["database", "Database"], ["functions", "Functions"], ["workflows", "Workflows"], ["files", "Files"], ["live-sessions", "Live sessions"], ["routes", "Routes"], ["monitoring", "Monitoring"], ["workers", "Workers"]] as const;
  return <aside className="sidebar"><p className="eyebrow">BRAIN CONTROL PLANE</p><h2>{title}</h2><nav className="sidebar-nav" aria-label="Admin navigation">{links.filter(([href]) => !visiblePages || visiblePages.includes(href)).map(([href, label]) => <Link className={activePage === href ? "active" : ""} href={`/${href}`} key={href}>{label}</Link>)}</nav>{projectID !== undefined && <label className="sidebar-project">Project<select value={projectID} disabled><option>{projectID || "—"}</option></select></label>}{onSignOut && <div className="sidebar-actions"><button className="secondary" onClick={onSignOut}>Sign out</button></div>}</aside>;
}

export function AdminHeader({ title, eyebrow = "BRAIN CONTROL PLANE" }: { title: string; eyebrow?: string }) {
  return <header className="admin-header"><p className="eyebrow">{eyebrow}</p><h1>{title}</h1></header>;
}

export function AdminFooter() {
  return <footer className="admin-footer">Brain Control Plane · Local administration</footer>;
}
