import { Suspense } from "react";
import { getTranslations } from "next-intl/server";
import { getApprovals, getProcedureDrafts } from "../lib/api";
import ApprovalForm from "../components/ApprovalForm";
import ApprovalListCard from "../components/ApprovalListCard";

async function ApprovalsContent() {
  const t = await getTranslations("approvals");
  const tc = await getTranslations("common");
  const [approvals, drafts] = await Promise.all([getApprovals(), getProcedureDrafts()]);

  const pendingApproval = approvals.items.find((a) => a.status === "waiting");

  return (
    <main className="shell">
      <header className="hero-page hero-compact">
        <p className="eyebrow">{t("eyebrow")}</p>
        <h1>{t("title")}</h1>
        <p className="lede">
          {t("lede")}
        </p>
      </header>

      {pendingApproval && <ApprovalForm approval={pendingApproval} />}

      <div className="panel-grid panel-grid-wide">
        <article className="panel">
          <div className="panel-header">
            <p className="section-kicker">{t("decisions")}</p>
            <h2 className="panel-title">{t("approvalHistory", { count: approvals.items.length })}</h2>
          </div>
          <div className="list">
            {approvals.items.length === 0 ? (
              <p className="empty-note">{t("noDecisions")}</p>
            ) : (
              approvals.items.map((approval) => (
                <ApprovalListCard key={approval.id} approval={approval} />
              ))
            )}
          </div>
        </article>

        <article className="panel">
          <div className="panel-header">
            <p className="section-kicker">{t("procedureDrafts")}</p>
            <h2 className="panel-title">{t("verificationQueue", { count: drafts.items.length })}</h2>
          </div>
          <div className="list">
            {drafts.items.length === 0 ? (
              <p className="empty-note">{t("noDrafts")}</p>
            ) : (
              drafts.items.map((draft) => (
                <article className="list-card" key={draft.id}>
                  <div className="list-card-main">
                    <h3 className="list-card-title">{draft.title}</h3>
                    <p className="list-card-meta">
                      {draft.source_type} · {draft.required_tools.join(", ")}
                    </p>
                  </div>
                  <span className={`badge badge-${draft.validation_status}`}>{draft.validation_status}</span>
                </article>
              ))
            )}
          </div>
        </article>
      </div>
    </main>
  );
}

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

      <div className="form-panel">
        <div className="loading-pulse loading-pulse-short" style={{ marginBottom: 16 }} />
        <div className="form-fields">
          <div className="loading-pulse loading-pulse-medium" />
          <div className="loading-pulse loading-pulse-long" />
        </div>
      </div>

      <div className="grid">
        <div className="panel">
          <div className="loading-pulse loading-pulse-medium" style={{ marginBottom: 16 }} />
          <div className="skeleton-grid">
            {[1, 2, 3].map((i) => <div key={i} className="skeleton-card" />)}
          </div>
        </div>
        <div className="panel">
          <div className="loading-pulse loading-pulse-medium" style={{ marginBottom: 16 }} />
          <div className="skeleton-grid">
            {[1, 2, 3].map((i) => <div key={i} className="skeleton-card" />)}
          </div>
        </div>
      </div>
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