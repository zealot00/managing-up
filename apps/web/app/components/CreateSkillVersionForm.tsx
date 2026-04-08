"use client";

import { useState, FormEvent } from "react";
import { createSkillVersion, Skill } from "../lib/api";
import { useTranslations } from "next-intl";
import { useApiMutation } from "../lib/use-mutations";

type Props = {
  skills: Skill[];
};

export default function CreateSkillVersionForm({ skills }: Props) {
  const t = useTranslations("skills");
  const tc = useTranslations("common");
  const [isOpen, setIsOpen] = useState(false);
  const [skillId, setSkillId] = useState("");
  const [version, setVersion] = useState("");
  const [changeSummary, setChangeSummary] = useState("");
  const [approvalRequired, setApprovalRequired] = useState(false);
  const [specYaml, setSpecYaml] = useState("");

  const createVersionMutation = useApiMutation(createSkillVersion, {
    queryKeysToInvalidate: [["skill-versions"]],
    successMessage: tc("success") + ": Version created",
    onSuccess: () => {
      setSkillId("");
      setVersion("");
      setChangeSummary("");
      setApprovalRequired(false);
      setSpecYaml("");
      setIsOpen(false);
    },
  });

  function handleSubmit(e: FormEvent) {
    e.preventDefault();
    createVersionMutation.mutate({
      skill_id: skillId,
      version,
      change_summary: changeSummary,
      approval_required: approvalRequired,
      spec_yaml: specYaml,
    });
  }

  return (
    <>
      <button onClick={() => setIsOpen(!isOpen)} className="trigger-btn">
        {isOpen ? tc("cancel") : t("newVersion")}
      </button>

      {isOpen && (
        <form onSubmit={handleSubmit} className="form-panel">
          <div className="panel-header">
            <p className="section-kicker">{t("versions")}</p>
            <h2>{t("createVersion")}</h2>
          </div>

          {createVersionMutation.error && <p className="form-error">{createVersionMutation.error.message}</p>}

          <div className="form-fields">
            <label className="form-label">
              {t("skill")}
              <select
                value={skillId}
                onChange={(e) => setSkillId(e.target.value)}
                required
                className="form-select"
              >
                <option value="">{t("selectSkill")}</option>
                {skills.map((s) => (
                  <option key={s.id} value={s.id}>
                    {s.name} ({s.owner_team})
                  </option>
                ))}
              </select>
            </label>

            <label className="form-label">
              {t("version")}
              <input
                type="text"
                value={version}
                onChange={(e) => setVersion(e.target.value)}
                placeholder={t("versionNumberPlaceholder")}
                required
                className="form-input"
              />
            </label>

            <label className="form-label">
              {t("changelog")}
              <input
                type="text"
                value={changeSummary}
                onChange={(e) => setChangeSummary(e.target.value)}
                placeholder={t("changelogPlaceholder")}
                required
                className="form-input"
              />
            </label>

            <label className="form-label" style={{ display: "flex", alignItems: "center", gap: 12 }}>
              <input
                type="checkbox"
                checked={approvalRequired}
                onChange={(e) => setApprovalRequired(e.target.checked)}
                style={{ width: 18, height: 18 }}
              />
              Require approval before execution
            </label>

            <label className="form-label">
              {t("yamlSpec")}
              <textarea
                value={specYaml}
                onChange={(e) => setSpecYaml(e.target.value)}
                placeholder={`name: my_skill\ndescription: Does something useful\nsteps:\n  - id: step1\n    name: First step\n    tool: execute_command\n    args:\n      command: echo "hello"`}
                rows={10}
                required
                className="form-textarea"
                style={{ fontFamily: "monospace", fontSize: "0.85rem" }}
              />
            </label>
          </div>

          <button type="submit" disabled={createVersionMutation.isPending} className="form-submit">
            {createVersionMutation.isPending ? t("creating") : t("createVersion")}
          </button>
        </form>
      )}
    </>
  );
}
