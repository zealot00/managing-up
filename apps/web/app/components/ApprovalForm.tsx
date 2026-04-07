"use client";

import { useState, FormEvent } from "react";
import { useRouter } from "next/navigation";
import { approveExecution, Approval } from "../lib/api";
import { useTranslations } from "next-intl";
import { useToast } from "../../components/ToastProvider";

type Props = {
  approval: Approval;
  onComplete?: () => void;
};

export default function ApprovalForm({ approval, onComplete }: Props) {
  const t = useTranslations("approvals");
  const tc = useTranslations("common");
  const router = useRouter();
  const toast = useToast();
  const [approver, setApprover] = useState("");
  const [note, setNote] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function handleDecision(decision: "approved" | "rejected") {
    if (!approver.trim()) {
      setError(t("approverPlaceholder"));
      return;
    }
    setLoading(true);
    setError("");

    try {
      await approveExecution(approval.execution_id, {
        approver,
        decision,
        note,
      });
      toast.success(tc("success") + ": Decision submitted");
      router.refresh();
      if (onComplete) onComplete();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to submit decision");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="form-panel">
      <div className="panel-header">
        <p className="section-kicker">{t("eyebrow")}</p>
        <h2>{t("decision", { skill_name: approval.skill_name })}</h2>
      </div>

      {error && <p className="form-error">{error}</p>}

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
          disabled={loading || !approver.trim()}
          className="btn-approve"
        >
          {loading ? t("submitting") : t("approve")}
        </button>
        <button
          type="button"
          onClick={() => handleDecision("rejected")}
          disabled={loading || !approver.trim()}
          className="btn-reject"
        >
          {loading ? t("submitting") : t("reject")}
        </button>
      </div>
    </div>
  );
}
