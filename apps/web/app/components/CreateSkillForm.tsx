"use client";

import { useState, FormEvent } from "react";
import { useRouter } from "next/navigation";
import { createSkill } from "../lib/api";
import { useTranslations } from "next-intl";
import { useToast } from "../../components/ToastProvider";

export default function CreateSkillForm() {
  const t = useTranslations("skills");
  const tc = useTranslations("common");
  const router = useRouter();
  const toast = useToast();
  const [name, setName] = useState("");
  const [ownerTeam, setOwnerTeam] = useState("");
  const [riskLevel, setRiskLevel] = useState("medium");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError("");

    try {
      await createSkill({ name, owner_team: ownerTeam, risk_level: riskLevel });
      setName("");
      setOwnerTeam("");
      setRiskLevel("medium");
      toast.success(tc("success") + ": Skill created");
      router.refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create skill");
    } finally {
      setLoading(false);
    }
  }

  return (
    <form onSubmit={handleSubmit} className="form-panel">
      <div className="panel-header">
        <p className="section-kicker">{t("eyebrow")}</p>
        <h2>{t("registerSkill")}</h2>
      </div>

      {error && <p className="form-error">{error}</p>}

      <div className="form-fields">
        <label className="form-label">
          {t("skillName")}
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder={t("skillNamePlaceholder")}
            required
            className="form-input"
          />
        </label>

        <label className="form-label">
          {t("ownerTeam")}
          <input
            type="text"
            value={ownerTeam}
            onChange={(e) => setOwnerTeam(e.target.value)}
            placeholder={t("ownerTeamPlaceholder")}
            required
            className="form-input"
          />
        </label>

        <label className="form-label">
          {t("riskLevel")}
          <select
            value={riskLevel}
            onChange={(e) => setRiskLevel(e.target.value)}
            className="form-select"
          >
            <option value="low">{t("low")}</option>
            <option value="medium">{t("medium")}</option>
            <option value="high">{t("high")}</option>
          </select>
        </label>
      </div>

      <button type="submit" disabled={loading} className="form-submit">
        {loading ? t("registering") : t("registerSkill")}
      </button>
    </form>
  );
}
