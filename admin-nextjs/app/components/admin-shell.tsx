"use client";
import { readAdminSession, signOutAdminSession } from "../lib/admin-api";

import { usePathname, useRouter } from "next/navigation";
import type { ReactNode } from "react";
import { useEffect, useState } from "react";
import { AdminFooter, AdminHeader, AdminSidebar } from "./admin-chrome";

const legacyShellRoutes = new Set(["/"]);
const pageTitles: Record<string, string> = {
  functions: "Functions",
  workflows: "Workflows",
  workers: "Workers",
};

export function AdminShell({ children }: { children: ReactNode }) {
  const pathname = usePathname();
  const router = useRouter();
  const [authenticated, setAuthenticated] = useState(false);
  useEffect(() => {
    setAuthenticated(Boolean(readAdminSession()));
  }, [pathname]);
  if (legacyShellRoutes.has(pathname) || !authenticated) return <>{children}</>;
  const pageName = pathname.split("/")[1] ?? "projects";
  return (
    <main className="shell">
      <div className="workspace">
        <AdminSidebar
          activePage={pageName}
          onSignOut={() => {
            void signOutAdminSession().finally(() => router.replace("/"));
          }}
        />
        <section className="card admin-content">
          <AdminHeader title={pageTitles[pageName] ?? pageName} />
          {children}
          <AdminFooter />
        </section>
      </div>
    </main>
  );
}
