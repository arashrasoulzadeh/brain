"use client";

import { ReactNode, useEffect, useState } from "react";

type ToastKind = "error" | "warning" | "info" | "success";
type ToastEvent = { kind: ToastKind; message: string };
type APIMessage = { status?: string; message?: string; code?: string };

function parseMessage(body: string, fallback: string): string {
  try {
    const parsed = JSON.parse(body) as APIMessage;
    return parsed.message || parsed.code || fallback;
  } catch {
    return body.trim() || fallback;
  }
}

export function showToast(kind: ToastKind, message: string): void {
  window.dispatchEvent(new CustomEvent<ToastEvent>("brain:toast", { detail: { kind, message } }));
}

export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<Array<ToastEvent & { id: number }>>([]);

  useEffect(() => {
    const onToast = (event: Event) => {
      const toast = (event as CustomEvent<ToastEvent>).detail;
      const id = Date.now() + Math.random();
      setToasts((current) => [...current, { ...toast, id }]);
      window.setTimeout(() => setToasts((current) => current.filter((item) => item.id !== id)), 5000);
    };
    const originalFetch = window.fetch.bind(window);
    window.fetch = async (...args) => {
      const response = await originalFetch(...args);
      if (response.url.includes("/api/")) {
        const body = await response.clone().text();
        if (!response.ok) showToast("error", parseMessage(body, `Request failed (${response.status})`));
        else if (body) {
          try {
            const parsed = JSON.parse(body) as APIMessage;
            if (parsed.status === "warning" || parsed.status === "info" || parsed.status === "success") {
              showToast(parsed.status, parsed.message || "Request completed.");
            }
          } catch {
            // Non-JSON success responses do not carry a toast message.
          }
        }
      }
      return response;
    };
    window.addEventListener("brain:toast", onToast);
    return () => {
      window.fetch = originalFetch;
      window.removeEventListener("brain:toast", onToast);
    };
  }, []);

  return <>{children}<div className="fixed right-6 top-6 z-50 flex w-80 max-w-[calc(100vw-3rem)] flex-col gap-3" aria-live="polite">{toasts.map((toast) => <div className={`rounded-xl border px-4 py-3 text-sm font-semibold shadow-xl ${toast.kind === "error" ? "border-red-200 bg-red-50 text-red-800" : toast.kind === "warning" ? "border-amber-200 bg-amber-50 text-amber-800" : toast.kind === "success" ? "border-emerald-200 bg-emerald-50 text-emerald-800" : "border-blue-200 bg-blue-50 text-blue-800"}`} key={toast.id}>{toast.message}</div>)}</div></>;
}
