"use client";

import { X, Trash2, CheckCircle, XCircle } from "lucide-react";

interface BulkAction {
  label: string;
  icon?: React.ReactNode;
  variant?: "primary" | "danger" | "secondary";
  onClick: () => void;
}

interface BulkActionBarProps {
  selectedCount: number;
  onClear: () => void;
  actions: BulkAction[];
}

export function BulkActionBar({ selectedCount, onClear, actions }: BulkActionBarProps) {
  if (selectedCount === 0) return null;

  return (
    <div className="bulk-action-bar">
      <div className="bulk-action-bar-info">
        <span className="bulk-action-count">{selectedCount} selected</span>
        <button type="button" onClick={onClear} className="bulk-action-clear">
          <X size={16} />
        </button>
      </div>
      <div className="bulk-action-bar-actions">
        {actions.map((action, i) => (
          <button
            key={i}
            type="button"
            onClick={action.onClick}
            className={`btn ${action.variant === "danger" ? "btn-danger" : action.variant === "primary" ? "btn-primary" : "btn-secondary"}`}
          >
            {action.icon}
            {action.label}
          </button>
        ))}
      </div>
    </div>
  );
}