"use client";

import { useState, FormEvent } from "react";
import { createSEHPolicy } from "../lib/seh-api";
import { useTranslations } from "next-intl";
import { useToast } from "../../components/ToastProvider";

type Props = {
  onCreated?: () => void;
};

export default function CreatePolicyForm({ onCreated }: Props) {
  const t = useTranslations("seh");
  const tc = useTranslations("common");
  const toast = useToast();
  const [name, setName] = useState("");
  const [requireProvenance, setRequireProvenance] = useState(false);
  const [requireApprovedForScore, setRequireApprovedForScore] = useState(false);
  const [minSourceDiversity, setMinSourceDiversity] = useState(2);
  const [minGoldenWeight, setMinGoldenWeight] = useState(0.5);
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
      setName("");
      setRequireProvenance(false);
      setRequireApprovedForScore(false);
      setMinSourceDiversity(2);
      setMinGoldenWeight(0.5);
      toast.success(tc("success") + ": Policy created");
      onCreated?.();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create policy");
    } finally {
      setLoading(false);
    }
  }

  return (
    <form onSubmit={handleSubmit} className="form-panel">
      <div className="panel-header">
        <p className="section-kicker">{t("eyebrow")}</p>
        <h2>{t("createPolicy")}</h2>
      </div>

      {error && <p className="form-error">{error}</p>}

      <div className="form-fields">
        <label className="form-label">
          {t("policyName")}
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder={t("policyNamePlaceholder")}
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

      <button type="submit" disabled={loading} className="form-submit">
        {loading ? t("creating") : t("createPolicy")}
      </button>
    </form>
  );
}