import { Suspense } from "react";
import SweepsPageClient from "./SweepsPageClient";

function Loading() {
  return (
    <div className="admin-content">
      <div className="loading-pulse loading-pulse-long" style={{ marginBottom: 16 }} />
      <div className="skeleton-grid">
        <div className="skeleton-card" />
        <div className="skeleton-card" />
      </div>
    </div>
  );
}

export default function SweepsPage() {
  return (
    <Suspense fallback={<Loading />}>
      <SweepsPageClient />
    </Suspense>
  );
}