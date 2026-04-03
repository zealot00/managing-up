"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { approveExecution, Approval } from "../lib/api";
import { useTranslations } from "next-intl";

type Props = {
  approval: Approval;
};

export default function ApprovalListCard({ approval }: Props) {
  const t = useTranslations("approvals");
  const router = useRouter();
  const [approver, setApprover] = useState("");
  const [note, setNote] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const isWaiting = approval.status === "waiting";

  async function handleDecision(decision: "approved" | "rejected") {
    if (!approver.trim()) {
      setError(t("approver"));
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
      router.refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to submit decision");
    } finally {
      setLoading(false);
    }
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
          {error && <p className="form-error">{error}</p>}
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
              disabled={loading || !approver.trim()}
              className="btn-approve"
            >
              {loading ? "..." : t("approve")}
            </button>
            <button
              type="button"
              onClick={() => handleDecision("rejected")}
              disabled={loading || !approver.trim()}
              className="btn-reject"
            >
              {loading ? "..." : t("reject")}
            </button>
          </div>
        </div>
      )}
    </article>
  );
}