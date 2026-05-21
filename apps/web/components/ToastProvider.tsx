"use client";

import { createContext, useContext, useState, useCallback, useRef, ReactNode } from "react";
import { X, CheckCircle, AlertCircle, AlertTriangle } from "lucide-react";

type ToastType = "success" | "error" | "warning" | "info";

interface Toast {
  id: string;
  message: string;
  type: ToastType;
  dismissing?: boolean;
}

interface ToastContextType {
  toast: (message: string, type?: ToastType) => void;
  success: (message: string) => void;
  error: (message: string) => void;
  warning: (message: string) => void;
  info: (message: string) => void;
}

const ToastContext = createContext<ToastContextType | undefined>(undefined);

export function useToast() {
  const context = useContext(ToastContext);
  if (!context) {
    throw new Error("useToast must be used within a ToastProvider");
  }
  return context;
}

function ToastIcon({ type }: { type: ToastType }) {
  switch (type) {
    case "success":
      return <CheckCircle size={18} aria-hidden="true" />;
    case "error":
      return <AlertCircle size={18} aria-hidden="true" />;
    case "warning":
      return <AlertTriangle size={18} aria-hidden="true" />;
    case "info":
    default:
      return <AlertCircle size={18} aria-hidden="true" />;
  }
}

function ToastItem({ toast, onDismiss }: { toast: Toast; onDismiss: (id: string) => void }) {
  // Error toasts use role="alert" (aria-live=assertive) for immediate screen reader announcement.
  // Other types inherit aria-live="polite" from the container.
  return (
    <div
      className={`toast toast-${toast.type}${toast.dismissing ? " toast-exit" : ""}`}
      role={toast.type === "error" ? "alert" : undefined}
    >
      <div className="toast-icon">
        <ToastIcon type={toast.type} />
      </div>
      <span className="toast-message">{toast.message}</span>
      <button
        className="toast-dismiss"
        onClick={() => onDismiss(toast.id)}
        aria-label="Dismiss notification"
      >
        <X size={16} aria-hidden="true" />
      </button>
    </div>
  );
}

export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<Toast[]>([]);
  const timersRef = useRef<Map<string, ReturnType<typeof setTimeout>>>(new Map());

  const dismissToast = useCallback((id: string) => {
    // Clear any pending auto-dismiss timer
    const timer = timersRef.current.get(id);
    if (timer) {
      clearTimeout(timer);
      timersRef.current.delete(id);
    }

    // Start exit animation
    setToasts((prev) =>
      prev.map((t) => (t.id === id ? { ...t, dismissing: true } : t)),
    );

    // Remove from DOM after animation completes
    setTimeout(() => {
      setToasts((prev) => prev.filter((t) => t.id !== id));
    }, 250);
  }, []);

  const addToast = useCallback((message: string, type: ToastType = "info") => {
    const id = `toast-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
    const newToast: Toast = { id, message, type };

    setToasts((prev) => [...prev, newToast]);

    // Auto-dismiss after 4 seconds
    const timer = setTimeout(() => {
      dismissToast(id);
    }, 4000);
    timersRef.current.set(id, timer);
  }, [dismissToast]);

  const toast = useCallback((message: string, type: ToastType = "info") => {
    addToast(message, type);
  }, [addToast]);

  const success = useCallback((message: string) => {
    addToast(message, "success");
  }, [addToast]);

  const error = useCallback((message: string) => {
    addToast(message, "error");
  }, [addToast]);

  const warning = useCallback((message: string) => {
    addToast(message, "warning");
  }, [addToast]);

  const info = useCallback((message: string) => {
    addToast(message, "info");
  }, [addToast]);

  return (
    <ToastContext.Provider value={{ toast, success, error, warning, info }}>
      {children}
      <div className="toast-container" aria-live="polite" aria-atomic="false">
        {toasts.map((t) => (
          <ToastItem key={t.id} toast={t} onDismiss={dismissToast} />
        ))}
      </div>
    </ToastContext.Provider>
  );
}
