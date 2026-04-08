"use client";

import { useState } from "react";
import { Approval, ProcedureDraft } from "../lib/api";
import { useTranslations } from "next-intl";
import ApprovalForm from "./ApprovalForm";
import { PageHeader } from "./layout/PageHeader";
import { EmptyState } from "./layout/EmptyState";
import { FormModal } from "./ui/FormModal";

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

      <FormModal
        isOpen={selectedApproval !== null}
        onClose={() => setSelectedApproval(null)}
        title={t("decision", { skill_name: selectedApproval?.skill_name || "" })}
        eyebrow={t("eyebrow")}
        error={undefined}
        isPending={false}
      >
        {selectedApproval && (
          <ApprovalForm
            approval={selectedApproval}
            onComplete={() => setSelectedApproval(null)}
          />
        )}
      </FormModal>
    </>
  );
}
