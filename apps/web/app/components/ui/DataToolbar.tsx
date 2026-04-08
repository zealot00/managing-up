"use client";

import { Search, X } from "lucide-react";

interface DataToolbarProps {
  searchQuery: string;
  onSearchChange: (value: string) => void;
  filters?: React.ReactNode;
  children?: React.ReactNode; // For action buttons on the right
}

export function DataToolbar({ searchQuery, onSearchChange, filters, children }: DataToolbarProps) {
  return (
    <div className="data-toolbar">
      <div className="data-toolbar-search">
        <Search size={18} className="search-icon" />
        <input
          type="text"
          value={searchQuery}
          onChange={(e) => onSearchChange(e.target.value)}
          placeholder="Search..."
          className="search-input"
        />
        {searchQuery && (
          <button
            type="button"
            onClick={() => onSearchChange("")}
            className="search-clear"
          >
            <X size={16} />
          </button>
        )}
      </div>
      
      {filters && <div className="data-toolbar-filters">{filters}</div>}
      
      {children && <div className="data-toolbar-actions">{children}</div>}
    </div>
  );
}