"use client";

import { useState, FormEvent } from "react";
import { createSEHPolicy } from "../lib/seh-api";
import { useTranslations } from "next-intl";

type Policy = {
  policy_id: string;
  name: string;
  require_provenance: boolean;
  require_approved_for_score?: boolean;
  min_source_diversity: number;
  min_golden_weight: number;
};

type Props = {
  policy: Policy;
  onCancel: () => void;
  onUpdated: () => void;
};

export default function EditPolicyForm({ policy, onCancel, onUpdated }: Props) {
  const t = useTranslations("seh");
  const [name, setName] = useState(policy.name);
  const [requireProvenance, setRequireProvenance] = useState(policy.require_provenance);
  const [requireApprovedForScore, setRequireApprovedForScore] = useState(policy.require_approved_for_score);
  const [minSourceDiversity, setMinSourceDiversity] = useState(policy.min_source_diversity);
  const [minGoldenWeight, setMinGoldenWeight] = useState(policy.min_golden_weight);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError("");

    try {
      await createSEHPolicy({
        name,
        require_provenance: requireProvenance,
        require_approved_for_score: requireApprovedForScore,
        min_source_diversity: minSourceDiversity,
        min_golden_weight: minGoldenWeight,
      });
      onUpdated();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to update policy");
    } finally {
      setLoading(false);
    }
  }

  return (
    <form onSubmit={handleSubmit} className="form-panel">
      <div className="panel-header">
        <p className="section-kicker">{t("governance")}</p>
        <h2>{t("editPolicy")}: {policy.name}</h2>
      </div>

      {error && <p className="form-error">{error}</p>}

      <div className="form-fields">
        <label className="form-label">
          {t("policyName")}
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
            className="form-input"
          />
        </label>

        <div className="checkbox-group">
          <label className="checkbox-label">
            <input
              type="checkbox"
              checked={requireProvenance}
              onChange={(e) => setRequireProvenance(e.target.checked)}
            />
            {t("requireProvenance")}
          </label>

          <label className="checkbox-label">
            <input
              type="checkbox"
              checked={requireApprovedForScore}
              onChange={(e) => setRequireApprovedForScore(e.target.checked)}
            />
            {t("requireApprovedForScore")}
          </label>
        </div>

        <div className="form-row">
          <label className="form-label">
            {t("minSourceDiversity")}
            <input
              type="number"
              value={minSourceDiversity}
              onChange={(e) => setMinSourceDiversity(Number(e.target.value))}
              min={1}
              className="form-input"
            />
          </label>

          <label className="form-label">
            {t("minGoldenWeight")}
            <input
              type="number"
              value={minGoldenWeight}
              onChange={(e) => setMinGoldenWeight(Number(e.target.value))}
              min={0}
              max={1}
              step={0.1}
              className="form-input"
            />
          </label>
        </div>
      </div>

      <div className="form-actions">
        <button type="submit" disabled={loading} className="form-submit" style={{ flex: 1 }}>
          {loading ? t("saving") : t("saveChanges")}
        </button>
        <button type="button" onClick={onCancel} className="btn btn-secondary" style={{ flex: 1 }}>
          {t("cancel") || "Cancel"}
        </button>
      </div>
    </form>
  );
}
