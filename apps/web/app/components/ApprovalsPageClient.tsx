"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Approval, ProcedureDraft, getApprovals, getProcedureDrafts, approveExecution } from "../lib/api";
import { useApiMutation } from "../lib/use-mutations";
import { useTranslations } from "next-intl";
import ApprovalForm from "./ApprovalForm";
import { PageHeader } from "./layout/PageHeader";
import { EmptyState } from "./layout/EmptyState";
import { FormModal } from "./ui/FormModal";
import { ListSkeleton } from "./layout/Skeleton";
import { BulkActionBar } from "./ui/BulkActionBar";
import { SelectableCard } from "./ui/SelectableCard";
import { CheckCircle, XCircle } from "lucide-react";

type TabType = "pending" | "drafts" | "history";

export default function ApprovalsPageClient() {
  const t = useTranslations("approvals");
  const [activeTab, setActiveTab] = useState<TabType>("pending");
  const [selectedApproval, setSelectedApproval] = useState<Approval | null>(null);
  const [selectedApprovalIds, setSelectedApprovalIds] = useState<Set<string>>(new Set());

  const { data: approvalsData, isLoading: isLoadingApprovals, isFetching: isFetchingApprovals } = useQuery({
    queryKey: ["approvals"],
    queryFn: getApprovals,
    placeholderData: (previousData) => previousData,
  });

  const { data: draftsData, isLoading: isLoadingDrafts, isFetching: isFetchingDrafts } = useQuery({
    queryKey: ["procedure-drafts"],
    queryFn: getProcedureDrafts,
    placeholderData: (previousData) => previousData,
  });

  const approvals = approvalsData ?? { items: [] };
  const drafts = draftsData ?? { items: [] };

  const isLoading = isLoadingApprovals && isLoadingDrafts;
  const isFetching = isFetchingApprovals || isFetchingDrafts;

  const pendingApprovals = approvals.items.filter((a) => a.status === "waiting");
  const completedApprovals = approvals.items.filter((a) => a.status !== "waiting");

  function toggleApprovalSelection(id: string) {
    setSelectedApprovalIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  }

  function clearSelection() {
    setSelectedApprovalIds(new Set());
  }

  const approveMutation = useApiMutation(
    ({ executionId, approver, decision, note }: { executionId: string; approver: string; decision: "approved" | "rejected"; note: string }) =>
      approveExecution(executionId, { approver, decision, note }),
    {
      queryKeysToInvalidate: [["approvals"], ["executions"]],
      successMessage: "",
    }
  );

  async function handleBulkApprove(decision: "approved" | "rejected") {
    const approver = prompt("Enter approver name:");
    if (!approver) return;
    const note = decision === "approved" ? "Bulk approved" : "Bulk rejected";
    for (const approval of pendingApprovals.filter((a) => selectedApprovalIds.has(a.id))) {
      await approveMutation.mutateAsync({
        executionId: approval.execution_id,
        approver,
        decision,
        note,
      });
    }
    clearSelection();
  }

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
          onClick={() => { setActiveTab("pending"); clearSelection(); }}
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
          onClick={() => { setActiveTab("drafts"); clearSelection(); }}
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
          onClick={() => { setActiveTab("history"); clearSelection(); }}
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
        {isLoading ? (
          <ListSkeleton rows={5} />
        ) : (
          <div style={{ opacity: isFetching && !isLoading ? 0.5 : 1, transition: "opacity 0.2s" }}>
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
                      <SelectableCard
                        key={approval.id}
                        isSelected={selectedApprovalIds.has(approval.id)}
                        onToggle={() => toggleApprovalSelection(approval.id)}
                      >
                        <article
                          className="list-card"
                          onClick={() => setSelectedApproval(approval)}
                          style={{ flex: 1, cursor: "pointer" }}
                        >
                          <div className="list-card-main">
                            <h3 className="list-card-title">{approval.skill_name}</h3>
                            <p className="list-card-meta">
                              {approval.step_id} · {approval.approver_group}
                            </p>
                          </div>
                          <span className={`badge badge-${approval.status}`}>{approval.status}</span>
                        </article>
                      </SelectableCard>
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
        )}
      </div>

      <BulkActionBar
        selectedCount={selectedApprovalIds.size}
        onClear={clearSelection}
        actions={[
          {
            label: `Approve (${selectedApprovalIds.size})`,
            icon: <CheckCircle size={16} />,
            variant: "primary",
            onClick: () => handleBulkApprove("approved"),
          },
          {
            label: `Reject (${selectedApprovalIds.size})`,
            icon: <XCircle size={16} />,
            variant: "danger",
            onClick: () => handleBulkApprove("rejected"),
          },
        ]}
      />

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