"use client";

import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { createSkill } from "../lib/api";
import { useTranslations } from "next-intl";
import { useApiMutation } from "../lib/use-mutations";
import { createSkillSchema } from "../lib/form-schemas";

export default function CreateSkillForm() {
  const t = useTranslations("skills");
  const tc = useTranslations("common");
  const createSkillMutation = useApiMutation(createSkill, {
    queryKeysToInvalidate: [["skills"]],
    successMessage: tc("success") + ": Skill created",
    onSuccess: () => {
      reset();
    },
  });

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting },
  } = useForm({
    resolver: zodResolver(createSkillSchema),
    defaultValues: {
      risk_level: "medium",
    },
  });

  function onSubmit(data: z.infer<typeof createSkillSchema>) {
    createSkillMutation.mutate({
      name: data.name,
      owner_team: data.owner_team,
      risk_level: data.risk_level,
    });
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="form-panel">
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
            {...register("name")}
            placeholder={t("skillNamePlaceholder")}
            className={`form-input ${errors.name ? "border-red-500" : ""}`}
          />
          {errors.name && <p className="form-error">{errors.name.message}</p>}
        </label>

        <label className="form-label">
          {t("ownerTeam")}
          <input
            type="text"
            {...register("owner_team")}
            placeholder={t("ownerTeamPlaceholder")}
            className={`form-input ${errors.owner_team ? "border-red-500" : ""}`}
          />
          {errors.owner_team && <p className="form-error">{errors.owner_team.message}</p>}
        </label>

        <label className="form-label">
          {t("riskLevel")}
          <select {...register("risk_level")} className="form-select">
            <option value="low">{t("low")}</option>
            <option value="medium">{t("medium")}</option>
            <option value="high">{t("high")}</option>
          </select>
          {errors.risk_level && <p className="form-error">{errors.risk_level.message}</p>}
        </label>
      </div>

      <button type="submit" disabled={isSubmitting} className="form-submit">
        {isSubmitting ? t("registering") : t("registerSkill")}
      </button>
    </form>
  );
}