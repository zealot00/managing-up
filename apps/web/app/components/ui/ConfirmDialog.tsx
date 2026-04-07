"use client";

import { useState, useEffect, ReactNode } from "react";
import { AlertTriangle, X } from "lucide-react";

interface ConfirmDialogProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: () => void;
  title: string;
  description?: string;
  confirmText?: string;
  cancelText?: string;
  variant?: "danger" | "warning" | "default";
  children?: ReactNode;
}

export function ConfirmDialog({
  isOpen,
  onClose,
  onConfirm,
  title,
  description,
  confirmText = "Confirm",
  cancelText = "Cancel",
  variant = "default",
  children,
}: ConfirmDialogProps) {
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (isOpen) {
      setLoading(false);
    }
  }, [isOpen]);

  async function handleConfirm() {
    setLoading(true);
    try {
      await onConfirm();
      onClose();
    } catch {
      setLoading(false);
    }
  }

  if (!isOpen) return null;

  const iconBgColor = variant === "danger"
    ? "var(--danger-bg)"
    : variant === "warning"
    ? "var(--warning-bg)"
    : "var(--neutral-bg)";

  const iconColor = variant === "danger"
    ? "var(--danger)"
    : variant === "warning"
    ? "var(--warning)"
    : "var(--neutral)";

  const buttonBgColor = variant === "danger"
    ? "var(--danger)"
    : variant === "warning"
    ? "var(--warning)"
    : "var(--primary)";

  return (
    <div
      style={{
        position: "fixed",
        inset: 0,
        background: "rgba(0, 0, 0, 0.5)",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        zIndex: 1000,
        padding: "var(--space-6)",
        animation: "fadeIn 0.15s ease-out",
      }}
      onClick={(e) => {
        if (e.target === e.currentTarget && !loading) onClose();
      }}
    >
      <div
        style={{
          background: "var(--surface-raised)",
          borderRadius: "var(--radius-lg)",
          padding: "var(--space-6)",
          width: "100%",
          maxWidth: 420,
          boxShadow: "var(--shadow-lg)",
          animation: "scaleIn 0.15s ease-out",
        }}
      >
        <div style={{ display: "flex", gap: "var(--space-4)", alignItems: "flex-start" }}>
          <div
            style={{
              width: 40,
              height: 40,
              borderRadius: "var(--radius-md)",
              background: iconBgColor,
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
              flexShrink: 0,
            }}
          >
            <AlertTriangle size={20} style={{ color: iconColor }} />
          </div>

          <div style={{ flex: 1, minWidth: 0 }}>
            <h3
              style={{
                fontSize: "var(--text-lg)",
                fontWeight: 700,
                color: "var(--ink-strong)",
                marginBottom: "var(--space-2)",
              }}
            >
              {title}
            </h3>
            {description && (
              <p style={{ fontSize: "var(--text-sm)", color: "var(--muted)", lineHeight: 1.6 }}>
                {description}
              </p>
            )}
            {children && (
              <div style={{ marginTop: "var(--space-4)" }}>
                {children}
              </div>
            )}
          </div>

          <button
            onClick={() => !loading && onClose()}
            disabled={loading}
            style={{
              background: "none",
              border: "none",
              padding: "var(--space-1)",
              cursor: loading ? "not-allowed" : "pointer",
              color: "var(--muted)",
              borderRadius: "var(--radius-sm)",
              transition: "all var(--transition-fast)",
            }}
          >
            <X size={18} />
          </button>
        </div>

        <div
          style={{
            display: "flex",
            gap: "var(--space-3)",
            marginTop: "var(--space-6)",
            justifyContent: "flex-end",
          }}
        >
          <button
            onClick={() => !loading && onClose()}
            disabled={loading}
            className="btn btn-secondary"
          >
            {cancelText}
          </button>
          <button
            onClick={handleConfirm}
            disabled={loading}
            style={{
              minHeight: 40,
              padding: "0 var(--space-5)",
              borderRadius: "var(--radius-sm)",
              background: buttonBgColor,
              color: "#ffffff",
              border: "none",
              fontSize: "var(--text-sm)",
              fontWeight: 600,
              cursor: loading ? "not-allowed" : "pointer",
              opacity: loading ? 0.7 : 1,
              transition: "all var(--transition-fast)",
            }}
          >
            {loading ? "..." : confirmText}
          </button>
        </div>
      </div>
    </div>
  );
}
