"use client";
import { readAdminSession, signOutAdminSession, systemUIURL, adminUIContext, ADMIN_SESSION_CHANGED } from "../lib/admin-api";

import { usePathname, useRouter } from "next/navigation";
import type { ReactNode } from "react";
import { useEffect, useState } from "react";
import { AdminFooter, AdminHeader, AdminSidebar, AdminHeadingContext, type PageHeading } from "./admin-chrome";

const legacyShellRoutes = new Set([systemUIURL()]);
const pageTitles: Record<string, string> = {
  dashboard: "Dashboard",
  functions: "Functions",
  workflows: "Workflows",
  workers: "Workers",
  projects: "Projects", plugins: "Plugins", audit: "Audit log", system: "System capabilities", routes: "API routes", monitoring: "Monitoring", database: "Database", files: "Files", "live-sessions": "Live sessions", extensions: "Plugin administration",
};

export function AdminShell({ children }: { children: ReactNode }) {
  const pathname = usePathname();
  const router = useRouter();
  const [authenticated, setAuthenticated] = useState(false);
  const [heading, setHeading] = useState<PageHeading>();
  useEffect(() => {
    const sync = () => {
      const token = readAdminSession();
      setAuthenticated(Boolean(token));
      if (!token && pathname !== systemUIURL()) router.replace(systemUIURL());
    };
    window.addEventListener(ADMIN_SESSION_CHANGED, sync);
    sync();
    return () => window.removeEventListener(ADMIN_SESSION_CHANGED, sync);
  }, [pathname, router]);
  if (legacyShellRoutes.has(pathname)) return <>{children}</>;
  if (!authenticated) return <main className="shell"><p role="status">Checking operator session…</p></main>;
  const { page: pageName, namespace } = adminUIContext(pathname);
  return (
    <AdminHeadingContext.Provider value={setHeading}>
    <main className="shell">
      <a className="skip-link" href="#admin-content">Skip to content</a>
      <div className="workspace">
        <AdminSidebar
          activePage={pageName}
          projectID={namespace}
          onSignOut={() => {
            void signOutAdminSession().finally(() => router.replace(systemUIURL()));
          }}
        />
        <section className="card admin-content" id="admin-content">
          <AdminHeader title={heading?.path === pathname ? heading.title : pageTitles[pageName] ?? pageName} eyebrow={heading?.path === pathname ? heading.eyebrow : undefined} />
          {children}
          <AdminFooter />
        </section>
      </div>
    </main>
    </AdminHeadingContext.Provider>
  );
}
