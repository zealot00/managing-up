"use client";

import { useState, useRef, useCallback } from "react";
import { useQuery } from "@tanstack/react-query";
import { Approval, ProcedureDraft, getApprovals, getProcedureDrafts, approveExecution } from "../lib/api";
import { useApiMutation } from "../lib/use-mutations";
import { useTranslations } from "next-intl";
import ApprovalForm from "./ApprovalForm";
import { PageHeader } from "./layout/PageHeader";
import { EmptyState } from "./layout/EmptyState";
import { QueryError } from "./layout/QueryError";
import { FormModal } from "./ui/FormModal";
import { RefreshIndicator } from "./ui/RefreshIndicator";
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
  const [bulkDecision, setBulkDecision] = useState<"approved" | "rejected" | null>(null);
  const [bulkApprover, setBulkApprover] = useState("");

  const { data: approvalsData, isLoading: isLoadingApprovals, isFetching: isFetchingApprovals, isError: isErrorApprovals, refetch: refetchApprovals } = useQuery({
    queryKey: ["approvals"],
    queryFn: getApprovals,
    placeholderData: (previousData) => previousData,
    refetchInterval: 10_000,
  });

  const { data: draftsData, isLoading: isLoadingDrafts, isFetching: isFetchingDrafts, isError: isErrorDrafts, refetch: refetchDrafts } = useQuery({
    queryKey: ["procedure-drafts"],
    queryFn: getProcedureDrafts,
    placeholderData: (previousData) => previousData,
    refetchInterval: 30_000,
  });

  const approvals = approvalsData ?? { items: [] };
  const drafts = draftsData ?? { items: [] };

  const isLoading = isLoadingApprovals && isLoadingDrafts;
  const isFetching = isFetchingApprovals || isFetchingDrafts;
  const isError = isErrorApprovals || isErrorDrafts;

  function handleRetry() {
    if (isErrorApprovals) refetchApprovals();
    if (isErrorDrafts) refetchDrafts();
  }

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
      successMessage: "Decision submitted",
    }
  );

  async function handleBulkApprove() {
    if (!bulkApprover.trim() || !bulkDecision) return;
    const note = bulkDecision === "approved" ? "Bulk approved" : "Bulk rejected";
    for (const approval of pendingApprovals.filter((a) => selectedApprovalIds.has(a.id))) {
      await approveMutation.mutateAsync({
        executionId: approval.execution_id,
        approver: bulkApprover.trim(),
        decision: bulkDecision,
        note,
      });
    }
    setBulkDecision(null);
    setBulkApprover("");
    clearSelection();
  }

  function handleApprovalUpdated() {
    setSelectedApproval(null);
  }

  const tabs = ["pending", "drafts", "history"] as const;
  const tabRefs = useRef<(HTMLButtonElement | null)[]>([]);

  const handleTabKeyDown = useCallback((e: React.KeyboardEvent) => {
    const idx = tabs.indexOf(activeTab);
    let nextIdx = idx;
    if (e.key === "ArrowRight") {
      nextIdx = (idx + 1) % tabs.length;
    } else if (e.key === "ArrowLeft") {
      nextIdx = (idx - 1 + tabs.length) % tabs.length;
    } else if (e.key === "Home") {
      nextIdx = 0;
    } else if (e.key === "End") {
      nextIdx = tabs.length - 1;
    } else {
      return;
    }
    e.preventDefault();
    const nextTab = tabs[nextIdx];
    setActiveTab(nextTab);
    clearSelection();
    tabRefs.current[nextIdx]?.focus();
  }, [activeTab]);

  const tabStyle = (isActive: boolean): React.CSSProperties => ({
    padding: "var(--space-3) var(--space-4)",
    border: "none",
    background: "none",
    cursor: "pointer",
    fontWeight: 600,
    fontSize: "var(--text-sm)",
    color: isActive ? "var(--ink-strong)" : "var(--muted)",
    borderBottom: isActive ? "2px solid var(--primary)" : "2px solid transparent",
    marginBottom: -1,
  });

  return (
    <>
      <PageHeader
        eyebrow={t("eyebrow")}
        title={<>{t("title")} <RefreshIndicator isFetching={isFetching} isLoading={isLoading} /></>}
        description={t("lede")}
      />

      <div role="tablist" aria-label="Approvals sections" style={{ marginBottom: "var(--space-6)", borderBottom: "1px solid var(--line)", display: "flex", gap: "var(--space-1)" }}>
        <button
          ref={(el) => { tabRefs.current[0] = el; }}
          role="tab"
          aria-selected={activeTab === "pending"}
          aria-controls="panel-pending"
          id="tab-pending"
          tabIndex={activeTab === "pending" ? 0 : -1}
          onClick={() => { setActiveTab("pending"); clearSelection(); }}
          onKeyDown={handleTabKeyDown}
          className={`tab-btn ${activeTab === "pending" ? "tab-btn-active" : ""}`}
          style={tabStyle(activeTab === "pending")}
        >
          Pending ({pendingApprovals.length})
        </button>
        <button
          ref={(el) => { tabRefs.current[1] = el; }}
          role="tab"
          aria-selected={activeTab === "drafts"}
          aria-controls="panel-drafts"
          id="tab-drafts"
          tabIndex={activeTab === "drafts" ? 0 : -1}
          onClick={() => { setActiveTab("drafts"); clearSelection(); }}
          onKeyDown={handleTabKeyDown}
          className={`tab-btn ${activeTab === "drafts" ? "tab-btn-active" : ""}`}
          style={tabStyle(activeTab === "drafts")}
        >
          Drafts ({drafts.items.length})
        </button>
        <button
          ref={(el) => { tabRefs.current[2] = el; }}
          role="tab"
          aria-selected={activeTab === "history"}
          aria-controls="panel-history"
          id="tab-history"
          tabIndex={activeTab === "history" ? 0 : -1}
          onClick={() => { setActiveTab("history"); clearSelection(); }}
          onKeyDown={handleTabKeyDown}
          className={`tab-btn ${activeTab === "history" ? "tab-btn-active" : ""}`}
          style={tabStyle(activeTab === "history")}
        >
          History ({completedApprovals.length})
        </button>
      </div>

      <div className="tab-content">
        {isLoading ? (
          <ListSkeleton rows={5} />
        ) : isError ? (
          <QueryError onRetry={handleRetry} />
        ) : (
          <div style={{ opacity: isFetching && !isLoading ? 0.5 : 1, transition: "opacity 0.2s" }}>
            {activeTab === "pending" && (
              <section role="tabpanel" id="panel-pending" aria-labelledby="tab-pending" className="panel">
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
              <section role="tabpanel" id="panel-drafts" aria-labelledby="tab-drafts" className="panel">
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
              <section role="tabpanel" id="panel-history" aria-labelledby="tab-history" className="panel">
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
            onClick: () => { setBulkDecision("approved"); setBulkApprover(""); },
          },
          {
            label: `Reject (${selectedApprovalIds.size})`,
            icon: <XCircle size={16} />,
            variant: "danger",
            onClick: () => { setBulkDecision("rejected"); setBulkApprover(""); },
          },
        ]}
      />

      <FormModal
        isOpen={bulkDecision !== null}
        onClose={() => setBulkDecision(null)}
        title={bulkDecision === "approved" ? `Approve ${selectedApprovalIds.size} items` : `Reject ${selectedApprovalIds.size} items`}
        eyebrow="Bulk action"
        isPending={approveMutation.isPending}
        submitText={bulkDecision === "approved" ? "Approve" : "Reject"}
        onSubmit={handleBulkApprove}
      >
        <div className="form-group">
          <label className="form-label" htmlFor="bulk-approver-name">Approver name</label>
          <input
            id="bulk-approver-name"
            className="form-input"
            value={bulkApprover}
            onChange={(e) => setBulkApprover(e.target.value)}
            placeholder="Enter your name"
            autoFocus
          />
        </div>
      </FormModal>

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