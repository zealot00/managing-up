import { Suspense } from "react";
import ApprovalsPageClient from "../components/ApprovalsPageClient";
import { ListSkeleton } from "../components/layout/Skeleton";

function SkeletonApprovalsPage() {
  return (
    <main className="shell">
      <header className="hero-page hero-compact">
        <p className="eyebrow">Human Control</p>
        <h1>Approvals and procedure validation workload.</h1>
        <p className="lede">
          Keep risk-bearing execution checkpoints and incoming SOP drafts inside a controlled review lane.
        </p>
      </header>

      <div style={{ marginBottom: "var(--space-6)", borderBottom: "1px solid var(--line)", display: "flex", gap: "var(--space-1)" }}>
        {["Pending", "Drafts", "History"].map((tab, i) => (
          <div key={tab} className="loading-pulse" style={{ width: 80, height: 40, marginBottom: -1, borderRadius: "var(--radius-sm) var(--radius-sm) 0 0" }} />
        ))}
      </div>

      <ListSkeleton rows={5} />
    </main>
  );
}

export default function ApprovalsPage() {
  return (
    <Suspense fallback={<SkeletonApprovalsPage />}>
      <ApprovalsPageClient />
    </Suspense>
  );
}
