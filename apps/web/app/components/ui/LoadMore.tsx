"use client";

interface LoadMoreProps {
  hasMore: boolean;
  isLoading: boolean;
  onLoadMore: () => void;
  label?: string;
}

export function LoadMore({ hasMore, isLoading, onLoadMore, label = "Load more" }: LoadMoreProps) {
  if (!hasMore) return null;

  return (
    <div style={{ textAlign: "center", padding: "var(--space-6)" }}>
      <button
        onClick={onLoadMore}
        disabled={isLoading}
        className="btn btn-secondary"
      >
        {isLoading ? "Loading..." : label}
      </button>
    </div>
  );
}