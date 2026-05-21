import { Suspense } from "react";
import ApprovalsPageClient from "../components/ApprovalsPageClient";
import { ListSkeleton } from "../components/layout/Skeleton";
import { PageHeader } from "../components/layout/PageHeader";

function SkeletonApprovalsPage() {
  return (
    <main className="shell">
      <PageHeader eyebrow="Human Control" title="Approvals" description="Keep risk-bearing execution checkpoints and incoming SOP drafts inside a controlled review lane." />

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
