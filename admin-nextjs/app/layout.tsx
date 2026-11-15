import "./styles.css";
import type { ReactNode } from "react";
import { ToastProvider } from "./components/toast-provider";
import { AdminShell } from "./components/admin-shell";
export default function Layout({ children }: { children: ReactNode }) {
  return (
    <html lang="en">
      <body><ToastProvider><AdminShell>{children}</AdminShell></ToastProvider></body>
    </html>
  );
}
