"use client";

import { ReactNode } from "react";
import { Check } from "lucide-react";

interface SelectableCardProps {
  children: ReactNode;
  isSelected: boolean;
  onToggle: () => void;
  className?: string;
}

export function SelectableCard({ children, isSelected, onToggle, className = "" }: SelectableCardProps) {
  return (
    <div
      className={`selectable-card ${isSelected ? "selected" : ""} ${className}`}
      onClick={onToggle}
      style={{ cursor: "pointer" }}
    >
      <div className="selectable-card-checkbox">
        {isSelected && <Check size={14} />}
      </div>
      <div className="selectable-card-content">
        {children}
      </div>
    </div>
  );
}