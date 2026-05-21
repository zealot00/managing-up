"use client";

import {ReactNode, useEffect, useRef, useCallback} from "react";
import {X} from "lucide-react";
import {useFocusTrap} from "../../hooks/use-focus-trap";

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
  onSubmit,
  children,
}: FormModalProps) {
  const titleId = useRef(`form-modal-title-${Math.random().toString(36).slice(2, 9)}`).current;
  const setContainerRef = useFocusTrap(isOpen);

  const handleClose = useCallback(() => {
    if (!isPending) onClose();
  }, [isPending, onClose]);

  useEffect(() => {
    function handleEsc(e: KeyboardEvent) {
      if (e.key === "Escape") handleClose();
    }
    if (isOpen) {
      document.addEventListener("keydown", handleEsc);
      document.body.style.overflow = "hidden";
    }
    return () => {
      document.removeEventListener("keydown", handleEsc);
      document.body.style.overflow = "";
    };
  }, [isOpen, handleClose]);

  if (!isOpen) return null;

  return (
    <div
      role="presentation"
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
        if (e.target === e.currentTarget) handleClose();
      }}
    >
      <div
        ref={setContainerRef}
        role="dialog"
        aria-modal="true"
        aria-labelledby={titleId}
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
            <h2 id={titleId} style={{ fontSize: "var(--text-xl)", fontWeight: 700, color: "var(--ink-strong)" }}>
              {title}
            </h2>
          </div>
          <button
            onClick={handleClose}
            disabled={isPending}
            aria-label="Close"
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
          <p className="form-error" role="alert" style={{ marginBottom: "var(--space-4)" }}>
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
