import { Suspense } from "react";
import { getApprovals, getProcedureDrafts } from "../lib/api";
import ApprovalForm from "../components/ApprovalForm";
import ApprovalListCard from "../components/ApprovalListCard";
import { SkeletonPanel } from "../components/SkeletonPanel";

function SkeletonApprovalsPage() {
  return (
    <main className="shell">
      <section className="hero-page hero-compact">
        <p className="eyebrow">Human Control</p>
        <h1>Approvals and procedure validation workload.</h1>
        <p className="lede">
          Keep risk-bearing execution checkpoints and incoming SOP drafts inside a controlled review
          lane.
        </p>
      </section>

      <div className="form-panel">
        <div className="loading-pulse loading-pulse-short" style={{ marginBottom: 12 }} />
        <div className="form-fields">
          <div className="loading-pulse loading-pulse-medium" />
          <div className="loading-pulse loading-pulse-long" />
        </div>
        <div style={{ display: "flex", gap: 12 }}>
          <div className="loading-pulse" style={{ flex: 1, height: 44, borderRadius: 999 }} />
          <div className="loading-pulse" style={{ flex: 1, height: 44, borderRadius: 999 }} />
        </div>
      </div>

      <section className="panel-grid panel-grid-wide">
        <SkeletonPanel height={320} />
        <SkeletonPanel height={320} />
      </section>
    </main>
  );
}

async function ApprovalsContent() {
  const [approvals, drafts] = await Promise.all([getApprovals(), getProcedureDrafts()]);

  const pendingApproval = approvals.items.find((a) => a.status === "waiting");

  return (
    <main className="shell">
      <section className="hero-page hero-compact">
        <p className="eyebrow">Human Control</p>
        <h1>Approvals and procedure validation workload.</h1>
        <p className="lede">
          Keep risk-bearing execution checkpoints and incoming SOP drafts inside a controlled review
          lane.
        </p>
      </section>

      {pendingApproval && <ApprovalForm approval={pendingApproval} />}

      <section className="panel-grid panel-grid-wide">
        <article className="panel">
          <div className="panel-header">
            <p className="section-kicker">Approvals</p>
            <h2>Decision history</h2>
          </div>
          <div className="list">
            {approvals.items.map((approval) => (
              <ApprovalListCard key={approval.id} approval={approval} />
            ))}
          </div>
        </article>

        <article className="panel">
          <div className="panel-header">
            <p className="section-kicker">Procedure Drafts</p>
            <h2>Validation queue</h2>
          </div>
          <div className="list">
            {drafts.items.map((draft) => (
              <article className="list-card" key={draft.id}>
                <div>
                  <h3>{draft.title}</h3>
                  <p>
                    {draft.source_type} · {draft.required_tools.join(", ")}
                  </p>
                </div>
                <span className={`badge badge-${draft.validation_status}`}>{draft.validation_status}</span>
              </article>
            ))}
          </div>
        </article>
      </section>
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
