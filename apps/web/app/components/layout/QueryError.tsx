"use client";

import { AlertCircle, RotateCw } from "lucide-react";

interface QueryErrorProps {
  message?: string;
  onRetry?: () => void;
}

export function QueryError({ message, onRetry }: QueryErrorProps) {
  return (
    <div className="empty-state">
      <div className="empty-state-icon" style={{ opacity: 0.5 }}>
        <AlertCircle size={48} />
      </div>
      <h3 className="empty-state-title">
        {message || "Failed to load data"}
      </h3>
      <p className="empty-state-description">
        An unexpected error occurred. Please try again.
      </p>
      {onRetry && (
        <div style={{ marginTop: "var(--space-5)" }}>
          <button
            className="btn btn-secondary"
            onClick={onRetry}
            style={{ display: "inline-flex", alignItems: "center", gap: "var(--space-2)" }}
          >
            <RotateCw size={16} />
            Retry
          </button>
        </div>
      )}
    </div>
  );
}
