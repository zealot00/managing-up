"use client";

import { useState, FormEvent } from "react";
import { useRouter } from "next/navigation";
import { approveExecution, Approval } from "../lib/api";

type Props = {
  approval: Approval;
};

export default function ApprovalForm({ approval }: Props) {
  const router = useRouter();
  const [approver, setApprover] = useState("");
  const [note, setNote] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function handleDecision(decision: "approved" | "rejected") {
    if (!approver.trim()) {
      setError("Approver name is required");
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
    <div className="form-panel">
      <div className="panel-header">
        <p className="section-kicker">Human Control</p>
        <h2>Decision for {approval.skill_name}</h2>
      </div>

      {error && <p className="form-error">{error}</p>}

      <div className="form-fields">
        <label className="form-label">
          Approver
          <input
            type="text"
            value={approver}
            onChange={(e) => setApprover(e.target.value)}
            placeholder="e.g. ops_manager"
            required
            className="form-input"
          />
        </label>

        <label className="form-label">
          Resolution note
          <textarea
            value={note}
            onChange={(e) => setNote(e.target.value)}
            placeholder="Optional note..."
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
          {loading ? "Submitting..." : "Approve"}
        </button>
        <button
          type="button"
          onClick={() => handleDecision("rejected")}
          disabled={loading || !approver.trim()}
          className="btn-reject"
        >
          {loading ? "Submitting..." : "Reject"}
        </button>
      </div>
    </div>
  );
}
