"use client";

interface SkeletonProps {
  width?: string | number;
  height?: string | number;
  borderRadius?: string;
  className?: string;
  style?: React.CSSProperties;
}

export function Skeleton({ width, height, borderRadius, className = "", style }: SkeletonProps) {
  return (
    <div
      className={`loading-pulse ${className}`}
      style={{
        width: width ?? "100%",
        height: height ?? 16,
        borderRadius: borderRadius ?? "var(--radius-sm)",
        flexShrink: 0,
        ...style,
      }}
    />
  );
}

interface ListSkeletonProps {
  rows?: number;
}

export function ListSkeleton({ rows = 5 }: ListSkeletonProps) {
  return (
    <div className="skeleton-list">
      {Array.from({ length: rows }).map((_, i) => (
        <div
          key={i}
          className="skeleton-card"
          style={{
            display: "flex",
            alignItems: "center",
            gap: "var(--space-4)",
            padding: "var(--space-5)",
          }}
        >
          <Skeleton width={32} height={32} borderRadius="var(--radius-full)" />
          <div style={{ flex: 1, display: "flex", flexDirection: "column", gap: "var(--space-2)" }}>
            <Skeleton width="60%" height={14} />
            <Skeleton width="40%" height={12} />
          </div>
          <Skeleton width={60} height={24} borderRadius="var(--radius-full)" />
        </div>
      ))}
    </div>
  );
}

interface CardGridSkeletonProps {
  count?: number;
  columns?: number;
}

export function CardGridSkeleton({ count = 6, columns = 3 }: CardGridSkeletonProps) {
  return (
    <div
      className="skeleton-grid"
      style={{
        display: "grid",
        gridTemplateColumns: `repeat(${columns}, minmax(0, 1fr))`,
        gap: "var(--space-5)",
      }}
    >
      {Array.from({ length: count }).map((_, i) => (
        <div key={i} className="skeleton-card" style={{ minHeight: 140 }} />
      ))}
    </div>
  );
}

interface TableSkeletonProps {
  rows?: number;
  columns?: number;
}

export function TableSkeleton({ rows = 5, columns = 4 }: TableSkeletonProps) {
  return (
    <div className="table-wrapper">
      <table className="table">
        <thead>
          <tr>
            {Array.from({ length: columns }).map((_, i) => (
              <th key={i}>
                <Skeleton width={80} height={12} />
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {Array.from({ length: rows }).map((_, rowIndex) => (
            <tr key={rowIndex}>
              {Array.from({ length: columns }).map((_, colIndex) => (
                <td key={colIndex}>
                  <Skeleton width={colIndex === 0 ? 120 : 80} height={14} />
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

interface PageSkeletonProps {
  showHeader?: boolean;
  headerActions?: boolean;
  filterBar?: boolean;
  content?: "list" | "cards" | "table";
  contentCount?: number;
}

export function PageSkeleton({
  showHeader = true,
  headerActions = true,
  filterBar = false,
  content = "list",
  contentCount = 5,
}: PageSkeletonProps) {
  return (
    <main className="shell">
      {showHeader && (
        <header
          className="hero-page hero-compact"
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "flex-start",
          }}
        >
          <div>
            <Skeleton width={80} height={12} style={{ marginBottom: "var(--space-3)" }} />
            <Skeleton width={280} height={28} style={{ marginBottom: "var(--space-3)" }} />
            <Skeleton width={400} height={16} />
          </div>
          {headerActions && <Skeleton width={140} height={40} borderRadius="var(--radius-sm)" />}
        </header>
      )}

      {filterBar && (
        <div style={{ display: "flex", gap: "var(--space-4)", marginBottom: "var(--space-6)" }}>
          <Skeleton width={200} height={40} />
          <Skeleton width={120} height={40} />
        </div>
      )}

      <div className="panel">
        <Skeleton width={120} height={16} style={{ marginBottom: "var(--space-4)" }} />
        {content === "list" && <ListSkeleton rows={contentCount} />}
        {content === "cards" && <CardGridSkeleton count={contentCount} />}
        {content === "table" && <TableSkeleton rows={contentCount} columns={4} />}
      </div>
    </main>
  );
}
