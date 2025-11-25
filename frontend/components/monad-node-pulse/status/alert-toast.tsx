"use client";

import { useEffect } from "react";
import type { AlertToast as AlertToastType } from "@/types/monad";

interface AlertToastProps {
  alert: AlertToastType;
  onDismiss: () => void;
}

const ALERT_COLORS: Record<
  AlertToastType["severity"],
  { border: string; glow: string }
> = {
  info: { border: "#85E6FF", glow: "0 0 20px rgba(133,230,255,0.35)" },
  warning: { border: "#F59E0B", glow: "0 0 20px rgba(245,158,11,0.35)" },
  critical: { border: "#EF4444", glow: "0 0 28px rgba(239,68,68,0.45)" },
};

export function AlertToast({ alert, onDismiss }: AlertToastProps) {
  useEffect(() => {
    const timeout = setTimeout(onDismiss, 6500);
    return () => clearTimeout(timeout);
  }, [alert.id, onDismiss]);

  const visual = ALERT_COLORS[alert.severity] ?? ALERT_COLORS.info;

  return (
    <aside
      className="alert-toast"
      style={{
        borderColor: visual.border,
        boxShadow: visual.glow,
      }}
      role="status"
      aria-live="assertive"
    >
      <div>
        <span className="text-label">{alert.title}</span>
        <p className="text-body">{alert.description}</p>
      </div>
      <button type="button" className="status-toggle ghost" onClick={onDismiss}>
        Dismiss
      </button>
    </aside>
  );
}
