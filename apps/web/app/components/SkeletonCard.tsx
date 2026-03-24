"use client";
export function SkeletonCard() {
  return (
    <article className="list-card">
      <div style={{ flex: 1 }}>
        <div className="loading-pulse loading-pulse-short" style={{ marginBottom: 8 }} />
        <div className="loading-pulse loading-pulse-medium" />
      </div>
      <div className="loading-pulse" style={{ width: 80, height: 32 }} />
    </article>
  );
}