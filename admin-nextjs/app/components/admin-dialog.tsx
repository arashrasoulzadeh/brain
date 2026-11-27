"use client";
import { useEffect, useId, useRef, type ReactNode } from "react";

/** Native dialog provides focus containment, inert backdrop and keyboard access. */
export function AdminDialog({ title, scope, busy = false, onClose, children }: {
  title: string; scope: string; busy?: boolean; onClose: () => void; children: ReactNode;
}) {
  const ref = useRef<HTMLDialogElement>(null);
  const titleID = useId();
  useEffect(() => {
    const dialog = ref.current;
    const trigger = document.activeElement;
    dialog?.showModal();
    return () => {
      dialog?.close();
      if (trigger instanceof HTMLElement && trigger.isConnected) trigger.focus();
    };
  }, []);
  return <dialog ref={ref} className="admin-dialog" aria-labelledby={titleID} onCancel={event => {
    event.preventDefault();
    if (!busy) onClose();
  }}>
    <div className="admin-dialog-header"><div><h2 id={titleID}>{title}</h2><p className="muted">{scope}</p></div><button type="button" className="secondary" aria-label="Close dialog" disabled={busy} onClick={onClose}>Close</button></div>
    <div className="admin-dialog-body" aria-busy={busy}>{children}</div>
  </dialog>;
}

export function AdminProgress({ label, percent }: { label: string; percent?: number }) {
  return <div className="admin-progress" role="status"><span>{label}{percent !== undefined ? " · " + Math.round(percent) + "%" : ""}</span><progress aria-label={label} max={100} value={percent} /></div>;
}
