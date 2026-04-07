import { Suspense } from "react";
import { getTranslations } from "next-intl/server";
import { getApprovals, getProcedureDrafts } from "../lib/api";
import ApprovalsPageClient from "../components/ApprovalsPageClient";

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

      <div className="panel">
        <div className="loading-pulse loading-pulse-medium" style={{ marginBottom: 16 }} />
        <div className="skeleton-grid">
          {[1, 2, 3].map((i) => (
            <div key={i} className="skeleton-card" />
          ))}
        </div>
      </div>
    </main>
  );
}

async function ApprovalsContent() {
  const [approvals, drafts] = await Promise.all([getApprovals(), getProcedureDrafts()]);

  return (
    <main className="shell">
      <ApprovalsPageClient approvals={approvals} drafts={drafts} />
    </main>
  );
}

export default function ApprovalsPage() {
  return (
    <Suspense fallback={<SkeletonApprovalsPage />}>
      <ApprovalsContent />
    </Suspense>
  );
}
