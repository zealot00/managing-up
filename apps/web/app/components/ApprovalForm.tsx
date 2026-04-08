"use client";

import { useState } from "react";
import { approveExecution, Approval } from "../lib/api";
import { useTranslations } from "next-intl";
import { useApiMutation } from "../lib/use-mutations";

type Props = {
  approval: Approval;
  onComplete?: () => void;
};

export default function ApprovalForm({ approval, onComplete }: Props) {
  const t = useTranslations("approvals");
  const tc = useTranslations("common");
  const [approver, setApprover] = useState("");
  const [note, setNote] = useState("");

  const approveExecutionMutation = useApiMutation(
    (body: { approver: string; decision: "approved" | "rejected"; note: string }) =>
      approveExecution(approval.execution_id, body),
    {
      successMessage: tc("success") + ": Decision submitted",
      queryKeysToInvalidate: [["approvals"], ["executions"]],
      onSuccess: () => onComplete?.(),
    }
  );

  function handleDecision(decision: "approved" | "rejected") {
    if (!approver.trim()) {
      return;
    }
    approveExecutionMutation.mutate({ approver, decision, note });
  }

  return (
    <div className="form-panel">
      <div className="panel-header">
        <p className="section-kicker">{t("eyebrow")}</p>
        <h2>{t("decision", { skill_name: approval.skill_name })}</h2>
      </div>

      {approveExecutionMutation.error && <p className="form-error">{approveExecutionMutation.error.message}</p>}

      <div className="form-fields">
        <label className="form-label">
          {t("approver")}
          <input
            type="text"
            value={approver}
            onChange={(e) => setApprover(e.target.value)}
            placeholder={t("approverPlaceholder")}
            required
            className="form-input"
          />
        </label>

        <label className="form-label">
          {t("resolutionNote")}
          <textarea
            value={note}
            onChange={(e) => setNote(e.target.value)}
            placeholder={t("notePlaceholder")}
            rows={2}
            className="form-textarea"
          />
        </label>
      </div>

      <div className="form-actions">
        <button
          type="button"
          onClick={() => handleDecision("approved")}
          disabled={approveExecutionMutation.isPending || !approver.trim()}
          className="btn-approve"
        >
          {approveExecutionMutation.isPending ? t("submitting") : t("approve")}
        </button>
        <button
          type="button"
          onClick={() => handleDecision("rejected")}
          disabled={approveExecutionMutation.isPending || !approver.trim()}
          className="btn-reject"
        >
          {approveExecutionMutation.isPending ? t("submitting") : t("reject")}
        </button>
      </div>
    </div>
  );
}
