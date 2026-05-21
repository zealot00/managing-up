"use client";

import { Loader2 } from "lucide-react";

interface RefreshIndicatorProps {
  isFetching: boolean;
  isLoading?: boolean;
}

export function RefreshIndicator({ isFetching, isLoading }: RefreshIndicatorProps) {
  if (!isFetching || isLoading) return null;
  return (
    <span className="refresh-indicator" aria-label="Refreshing data">
      <Loader2 size={14} className="refresh-indicator-icon" />
    </span>
  );
}
