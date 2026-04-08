"use client";

import { useState } from "react";
import { approveExecution, Approval } from "../lib/api";
import { useTranslations } from "next-intl";
import { useApiMutation } from "../lib/use-mutations";

type Props = {
  approval: Approval;
};

export default function ApprovalListCard({ approval }: Props) {
  const t = useTranslations("approvals");
  const [approver, setApprover] = useState("");
  const [note, setNote] = useState("");

  const isWaiting = approval.status === "waiting";

  const approveExecutionMutation = useApiMutation(
    (body: { approver: string; decision: "approved" | "rejected"; note: string }) =>
      approveExecution(approval.execution_id, body),
    {
      queryKeysToInvalidate: [["approvals"], ["executions"]],
    }
  );

  function handleDecision(decision: "approved" | "rejected") {
    if (!approver.trim()) {
      return;
    }
    approveExecutionMutation.mutate({ approver, decision, note });
  }

  return (
    <article className="list-card">
      <div>
        <h3>{approval.skill_name}</h3>
        <p>
          {approval.step_id} · {approval.approver_group}
        </p>
      </div>
      <span className={`badge badge-${approval.status}`}>{approval.status}</span>

      {isWaiting && (
        <div className="inline-form">
          {approveExecutionMutation.error && <p className="form-error">{approveExecutionMutation.error.message}</p>}
          <input
            type="text"
            value={approver}
            onChange={(e) => setApprover(e.target.value)}
            placeholder={t("approverPlaceholder")}
            required
            className="form-input"
          />
          <textarea
            value={note}
            onChange={(e) => setNote(e.target.value)}
            placeholder={t("notePlaceholder")}
            rows={1}
            className="form-textarea"
          />
          <div className="form-actions">
            <button
              type="button"
              onClick={() => handleDecision("approved")}
              disabled={approveExecutionMutation.isPending || !approver.trim()}
              className="btn-approve"
            >
              {approveExecutionMutation.isPending ? "..." : t("approve")}
            </button>
            <button
              type="button"
              onClick={() => handleDecision("rejected")}
              disabled={approveExecutionMutation.isPending || !approver.trim()}
              className="btn-reject"
            >
              {approveExecutionMutation.isPending ? "..." : t("reject")}
            </button>
          </div>
        </div>
      )}
    </article>
  );
}