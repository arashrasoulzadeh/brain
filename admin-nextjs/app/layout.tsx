import "./styles.css";
import type { ReactNode } from "react";
import { ToastProvider } from "./components/toast-provider";
export default function Layout({ children }: { children: ReactNode }) {
  return (
    <html lang="en">
      <body><ToastProvider>{children}</ToastProvider></body>
    </html>
  );
}
