"use client";

import { useEffect, ReactNode } from "react";
import { X } from "lucide-react";

interface DrawerProps {
  isOpen: boolean;
  onClose: () => void;
  title: string;
  description?: string;
  children: ReactNode;
}

export function Drawer({ isOpen, onClose, title, description, children }: DrawerProps) {
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
        zIndex: 1000,
        display: "flex",
        justifyContent: "flex-end",
      }}
    >
      <div
        style={{
          position: "absolute",
          inset: 0,
          background: "rgba(0, 0, 0, 0.4)",
          animation: "fadeIn 0.2s ease-out",
        }}
        onClick={onClose}
      />
      <aside
        style={{
          position: "relative",
          width: "100%",
          maxWidth: 560,
          height: "100%",
          background: "var(--surface-raised)",
          boxShadow: "var(--shadow-lg)",
          display: "flex",
          flexDirection: "column",
          animation: "slideInRight 0.25s ease-out",
        }}
      >
        <div style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "flex-start",
          padding: "var(--space-6)",
          borderBottom: "1px solid var(--line)",
        }}>
          <div>
            <h2 style={{ fontSize: "var(--text-xl)", fontWeight: 700, color: "var(--ink-strong)", marginBottom: "var(--space-1)" }}>
              {title}
            </h2>
            {description && (
              <p style={{ fontSize: "var(--text-sm)", color: "var(--muted)" }}>{description}</p>
            )}
          </div>
          <button
            onClick={onClose}
            style={{
              background: "none",
              border: "none",
              padding: "var(--space-2)",
              cursor: "pointer",
              color: "var(--muted)",
              borderRadius: "var(--radius-sm)",
            }}
          >
            <X size={20} />
          </button>
        </div>
        <div style={{ flex: 1, overflow: "auto", padding: "var(--space-6)" }}>
          {children}
        </div>
      </aside>

      <style jsx global>{`
        @keyframes slideInRight {
          from { transform: translateX(100%); }
          to { transform: translateX(0); }
        }
        @keyframes fadeIn {
          from { opacity: 0; }
          to { opacity: 1; }
        }
      `}</style>
    </div>
  );
}