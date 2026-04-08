"use client";

import { useState } from "react";
import { Approval, ProcedureDraft } from "../lib/api";
import { useTranslations } from "next-intl";
import ApprovalForm from "./ApprovalForm";
import { PageHeader } from "./layout/PageHeader";
import { EmptyState } from "./layout/EmptyState";

type Props = {
  approvals: { items: Approval[] };
  drafts: { items: ProcedureDraft[] };
};

type TabType = "pending" | "drafts" | "history";

export default function ApprovalsPageClient({ approvals, drafts }: Props) {
  const t = useTranslations("approvals");
  const [activeTab, setActiveTab] = useState<TabType>("pending");
  const [selectedApproval, setSelectedApproval] = useState<Approval | null>(null);

  const pendingApprovals = approvals.items.filter((a) => a.status === "waiting");
  const completedApprovals = approvals.items.filter((a) => a.status !== "waiting");

  function handleApprovalUpdated() {
    setSelectedApproval(null);
  }

  return (
    <>
      <PageHeader
        eyebrow={t("eyebrow")}
        title={t("title")}
        description={t("lede")}
      />

      <div className="tabs" style={{ marginBottom: "var(--space-6)", borderBottom: "1px solid var(--line)", display: "flex", gap: "var(--space-1)" }}>
        <button
          onClick={() => setActiveTab("pending")}
          className={`tab-btn ${activeTab === "pending" ? "tab-btn-active" : ""}`}
          style={{
            padding: "var(--space-3) var(--space-4)",
            border: "none",
            background: "none",
            cursor: "pointer",
            fontWeight: 600,
            fontSize: "var(--text-sm)",
            color: activeTab === "pending" ? "var(--ink-strong)" : "var(--muted)",
            borderBottom: activeTab === "pending" ? "2px solid var(--primary)" : "2px solid transparent",
            marginBottom: -1,
          }}
        >
          Pending ({pendingApprovals.length})
        </button>
        <button
          onClick={() => setActiveTab("drafts")}
          className={`tab-btn ${activeTab === "drafts" ? "tab-btn-active" : ""}`}
          style={{
            padding: "var(--space-3) var(--space-4)",
            border: "none",
            background: "none",
            cursor: "pointer",
            fontWeight: 600,
            fontSize: "var(--text-sm)",
            color: activeTab === "drafts" ? "var(--ink-strong)" : "var(--muted)",
            borderBottom: activeTab === "drafts" ? "2px solid var(--primary)" : "2px solid transparent",
            marginBottom: -1,
          }}
        >
          Drafts ({drafts.items.length})
        </button>
        <button
          onClick={() => setActiveTab("history")}
          className={`tab-btn ${activeTab === "history" ? "tab-btn-active" : ""}`}
          style={{
            padding: "var(--space-3) var(--space-4)",
            border: "none",
            background: "none",
            cursor: "pointer",
            fontWeight: 600,
            fontSize: "var(--text-sm)",
            color: activeTab === "history" ? "var(--ink-strong)" : "var(--muted)",
            borderBottom: activeTab === "history" ? "2px solid var(--primary)" : "2px solid transparent",
            marginBottom: -1,
          }}
        >
          History ({completedApprovals.length})
        </button>
      </div>

      <div className="tab-content">
        {activeTab === "pending" && (
          <section className="panel">
            <div className="panel-header">
              <p className="section-kicker">{t("eyebrow")}</p>
              <h2 className="panel-title">{t("decisions")}</h2>
            </div>
            {pendingApprovals.length === 0 ? (
              <EmptyState title={t("noDecisions")} />
            ) : (
              <div className="list">
                {pendingApprovals.map((approval) => (
                  <article
                    className="list-card"
                    key={approval.id}
                    onClick={() => setSelectedApproval(approval)}
                    style={{ cursor: "pointer" }}
                  >
                    <div className="list-card-main">
                      <h3 className="list-card-title">{approval.skill_name}</h3>
                      <p className="list-card-meta">
                        {approval.step_id} · {approval.approver_group}
                      </p>
                    </div>
                    <span className={`badge badge-${approval.status}`}>{approval.status}</span>
                  </article>
                ))}
              </div>
            )}
          </section>
        )}

        {activeTab === "drafts" && (
          <section className="panel">
            <div className="panel-header">
              <p className="section-kicker">{t("procedureDrafts")}</p>
              <h2 className="panel-title">{t("verificationQueue", { count: drafts.items.length })}</h2>
            </div>
            {drafts.items.length === 0 ? (
              <EmptyState title={t("noDrafts")} />
            ) : (
              <div className="list">
                {drafts.items.map((draft) => (
                  <article className="list-card" key={draft.id}>
                    <div className="list-card-main">
                      <h3 className="list-card-title">{draft.title}</h3>
                      <p className="list-card-meta">
                        {draft.source_type} · {draft.required_tools.join(", ")}
                      </p>
                    </div>
                    <span className={`badge badge-${draft.validation_status}`}>{draft.validation_status}</span>
                  </article>
                ))}
              </div>
            )}
          </section>
        )}

        {activeTab === "history" && (
          <section className="panel">
            <div className="panel-header">
              <p className="section-kicker">{t("decisions")}</p>
              <h2 className="panel-title">{t("approvalHistory", { count: completedApprovals.length })}</h2>
            </div>
            {completedApprovals.length === 0 ? (
              <EmptyState title={t("noDecisions")} />
            ) : (
              <div className="list">
                {completedApprovals.map((approval) => (
                  <article className="list-card" key={approval.id}>
                    <div className="list-card-main">
                      <h3 className="list-card-title">{approval.skill_name}</h3>
                      <p className="list-card-meta">
                        {approval.step_id} · {approval.approver_group}
                      </p>
                    </div>
                    <span className={`badge badge-${approval.status}`}>{approval.status}</span>
                  </article>
                ))}
              </div>
            )}
          </section>
        )}
      </div>

      {selectedApproval && (
        <div
          style={{
            position: "fixed",
            inset: 0,
            background: "rgba(0, 0, 0, 0.5)",
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            zIndex: 1000,
            padding: "var(--space-6)",
          }}
          onClick={(e) => {
            if (e.target === e.currentTarget) setSelectedApproval(null);
          }}
        >
          <div
            style={{
              background: "var(--surface-raised)",
              borderRadius: "var(--radius-lg)",
              padding: "var(--space-6)",
              width: "100%",
              maxWidth: 520,
              maxHeight: "90vh",
              overflowY: "auto",
              boxShadow: "var(--shadow-lg)",
            }}
          >
            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: "var(--space-5)" }}>
              <div>
                <p className="section-kicker">{t("eyebrow")}</p>
                <h2 style={{ fontSize: "var(--text-xl)", fontWeight: 700, color: "var(--ink-strong)" }}>
                  {t("decision", { skill_name: selectedApproval.skill_name })}
                </h2>
              </div>
              <button
                onClick={() => setSelectedApproval(null)}
                style={{
                  background: "none",
                  border: "none",
                  fontSize: "var(--text-xl)",
                  cursor: "pointer",
                  color: "var(--muted)",
                  padding: "var(--space-2)",
                }}
              >
                ×
              </button>
            </div>

            <ApprovalForm approval={selectedApproval} onComplete={handleApprovalUpdated} />
          </div>
        </div>
      )}
    </>
  );
}
