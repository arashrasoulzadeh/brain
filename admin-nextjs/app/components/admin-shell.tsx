"use client";
import { readAdminSession, signOutAdminSession } from "../lib/admin-api";

import { usePathname, useRouter } from "next/navigation";
import type { ReactNode } from "react";
import { useEffect, useState } from "react";
import { AdminFooter, AdminHeader, AdminSidebar, AdminHeadingContext, type PageHeading } from "./admin-chrome";

const legacyShellRoutes = new Set(["/"]);
const pageTitles: Record<string, string> = {
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
    setAuthenticated(Boolean(readAdminSession()));
  }, [pathname]);
  if (legacyShellRoutes.has(pathname) || !authenticated) return <>{children}</>;
  const pageName = pathname.split("/")[1] ?? "projects";
  return (
    <AdminHeadingContext.Provider value={setHeading}>
    <main className="shell">
      <div className="workspace">
        <AdminSidebar
          activePage={pageName}
          onSignOut={() => {
            void signOutAdminSession().finally(() => router.replace("/"));
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
