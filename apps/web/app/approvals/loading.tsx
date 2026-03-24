export default function Loading() {
  return (
    <div className="loading-shell">
      <div className="loading-pulse loading-pulse-long" style={{ marginBottom: 16 }} />
      <div className="loading-pulse loading-pulse-medium" />
      <div className="skeleton-grid">
        <div className="skeleton-card" />
        <div className="skeleton-card" />
      </div>
    </div>
  );
}
