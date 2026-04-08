"use client";

import { useState } from "react";
import Link from "next/link";
import { useTranslations } from "next-intl";
import { Skill, createSkill } from "../lib/api";
import { useApiMutation } from "../lib/use-mutations";
import { PageHeader } from "./layout/PageHeader";
import { EmptyState } from "./layout/EmptyState";
import { FormModal } from "./ui/FormModal";

type Props = {
  skills: { items: Skill[] };
};

export default function SkillsPageClient({ skills }: Props) {
  const t = useTranslations("skills");
  const tc = useTranslations("common");
  const [showCreateModal, setShowCreateModal] = useState(false);

  const createSkillMutation = useApiMutation(createSkill, {
    queryKeysToInvalidate: [["skills"]],
    successMessage: tc("success") + ": Skill created",
    onSuccess: () => setShowCreateModal(false),
  });

  function handleCreateSkill(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();

    const formData = new FormData(e.currentTarget);
    const name = formData.get("name") as string;
    const ownerTeam = formData.get("owner_team") as string;
    const riskLevel = formData.get("risk_level") as string;

    createSkillMutation.mutate({ name, owner_team: ownerTeam, risk_level: riskLevel });
  }

  return (
    <>
      <PageHeader
        eyebrow={t("eyebrow")}
        title={t("title")}
        description={t("lede")}
        actions={
          <button
            className="btn btn-primary"
            onClick={() => setShowCreateModal(true)}
          >
            + {t("registerSkill")}
          </button>
        }
      />

      <div className="panel">
        <div className="panel-header">
          <p className="section-kicker">{t("eyebrow")}</p>
          <h2 className="panel-title">{t("title")}</h2>
        </div>
        <div className="table-wrapper">
          {skills.items.length === 0 ? (
            <EmptyState title={t("noSkills")} />
          ) : (
            <table className="table">
              <thead>
                <tr>
                  <th>Name</th>
                  <th>Owner</th>
                  <th>Risk Level</th>
                  <th>Version</th>
                  <th>Status</th>
                </tr>
              </thead>
              <tbody>
                {skills.items.map((skill) => (
                  <tr key={skill.id} style={{ cursor: "pointer" }}>
                    <td>
                      <Link href={`/skills/${skill.id}`} style={{ textDecoration: "none" }}>
                        {skill.name}
                      </Link>
                    </td>
                    <td>{skill.owner_team}</td>
                    <td>{skill.risk_level}</td>
                    <td>{skill.current_version || t("noVersions")}</td>
                    <td>
                      <span className={`badge badge-${skill.status}`}>{skill.status}</span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      </div>

      <FormModal
        isOpen={showCreateModal}
        onClose={() => setShowCreateModal(false)}
        title={t("registerSkill")}
        eyebrow={t("eyebrow")}
        error={createSkillMutation.isError ? createSkillMutation.error?.message : undefined}
        isPending={createSkillMutation.isPending}
      >
        <form onSubmit={handleCreateSkill}>
          <div className="form-fields">
            <label className="form-label">
              {t("skillName")}
              <input
                type="text"
                name="name"
                placeholder={t("skillNamePlaceholder")}
                required
                className="form-input"
              />
            </label>

            <label className="form-label">
              {t("ownerTeam")}
              <input
                type="text"
                name="owner_team"
                placeholder={t("ownerTeamPlaceholder")}
                required
                className="form-input"
              />
            </label>

            <label className="form-label">
              {t("riskLevel")}
              <select name="risk_level" className="form-select" defaultValue="medium">
                <option value="low">{t("low")}</option>
                <option value="medium">{t("medium")}</option>
                <option value="high">{t("high")}</option>
              </select>
            </label>
          </div>

          <button type="submit" disabled={createSkillMutation.isPending} className="form-submit" style={{ marginTop: "var(--space-4)" }}>
            {createSkillMutation.isPending ? t("registering") : t("registerSkill")}
          </button>
        </form>
      </FormModal>
    </>
  );
}
