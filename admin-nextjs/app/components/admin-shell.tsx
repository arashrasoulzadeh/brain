"use client";

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
    setAuthenticated(Boolean(window.sessionStorage.getItem("brain-admin-session")));
  }, [pathname]);
  if (legacyShellRoutes.has(pathname) || !authenticated) return <>{children}</>;
  const pageName = pathname.split("/")[1] ?? "projects";
  return (
    <main className="shell">
      <div className="workspace">
        <AdminSidebar
          activePage={pageName}
          onSignOut={() => {
            window.sessionStorage.removeItem("brain-admin-session");
            router.replace("/");
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
