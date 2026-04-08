"use client";

import {ReactNode, useEffect} from "react";
import {X} from "lucide-react";

interface FormModalProps {
  isOpen: boolean;
  onClose: () => void;
  title: string;
  eyebrow?: string;
  error?: string;
  isPending?: boolean;
  submitText?: string;
  onSubmit?: () => void;
  children: ReactNode;
}

export function FormModal({
  isOpen,
  onClose,
  title,
  eyebrow,
  error,
  isPending,
  submitText,
  children,
}: FormModalProps) {
  useEffect(() => {
    function handleEsc(e: KeyboardEvent) {
      if (e.key === "Escape") onClose();
    }
    if (isOpen) {
      document.addEventListener("keydown", handleEsc);
      document.body.style.overflow = "hidden";
    }
    return () => {
      document.removeEventListener("keydown", handleEsc);
      document.body.style.overflow = "";
    };
  }, [isOpen, onClose]);

  if (!isOpen) return null;

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
        if (e.target === e.currentTarget) onClose();
      }}
    >
      <div
        style={{
          background: "var(--surface-raised)",
          borderRadius: "var(--radius-lg)",
          padding: "var(--space-6)",
          width: "100%",
          maxWidth: 480,
          maxHeight: "90vh",
          overflowY: "auto",
          boxShadow: "var(--shadow-lg)",
          animation: "scaleIn 0.15s ease-out",
        }}
      >
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", marginBottom: "var(--space-5)" }}>
          <div>
            {eyebrow && <p className="section-kicker">{eyebrow}</p>}
            <h2 style={{ fontSize: "var(--text-xl)", fontWeight: 700, color: "var(--ink-strong)" }}>
              {title}
            </h2>
          </div>
          <button
            onClick={onClose}
            disabled={isPending}
            style={{
              background: "none",
              border: "none",
              fontSize: "var(--text-xl)",
              cursor: isPending ? "not-allowed" : "pointer",
              color: "var(--muted)",
              padding: "var(--space-2)",
            }}
          >
            ×
          </button>
        </div>

        {error && (
          <p className="form-error" style={{ marginBottom: "var(--space-4)" }}>
            {error}
          </p>
        )}

        {children}
      </div>

      <style jsx global>{`
        @keyframes fadeIn {
          from { opacity: 0; }
          to { opacity: 1; }
        }
        @keyframes scaleIn {
          from { opacity: 0; transform: scale(0.95); }
          to { opacity: 1; transform: scale(1); }
        }
      `}</style>
    </div>
  );
}