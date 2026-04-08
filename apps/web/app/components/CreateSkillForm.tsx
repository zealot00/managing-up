"use client";

import { useState, FormEvent } from "react";
import { createSkill } from "../lib/api";
import { useTranslations } from "next-intl";
import { useApiMutation } from "../lib/use-mutations";

export default function CreateSkillForm() {
  const t = useTranslations("skills");
  const tc = useTranslations("common");
  const [name, setName] = useState("");
  const [ownerTeam, setOwnerTeam] = useState("");
  const [riskLevel, setRiskLevel] = useState("medium");

  const createSkillMutation = useApiMutation(createSkill, {
    queryKeysToInvalidate: [["skills"]],
    successMessage: tc("success") + ": Skill created",
    onSuccess: () => {
      setName("");
      setOwnerTeam("");
      setRiskLevel("medium");
    },
  });

  function handleSubmit(e: FormEvent) {
    e.preventDefault();
    createSkillMutation.mutate({ name, owner_team: ownerTeam, risk_level: riskLevel });
  }

  return (
    <form onSubmit={handleSubmit} className="form-panel">
      <div className="panel-header">
        <p className="section-kicker">{t("eyebrow")}</p>
        <h2>{t("registerSkill")}</h2>
      </div>

      {createSkillMutation.error && <p className="form-error">{createSkillMutation.error.message}</p>}

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

      <button type="submit" disabled={createSkillMutation.isPending} className="form-submit">
        {createSkillMutation.isPending ? t("registering") : t("registerSkill")}
      </button>
    </form>
  );
}
