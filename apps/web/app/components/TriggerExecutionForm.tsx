"use client";

import { useState, FormEvent } from "react";
import { createExecution, Skill } from "../lib/api";
import { useApiMutation } from "../lib/use-mutations";
import { useTranslations } from "next-intl";
import { useToast } from "../../components/ToastProvider";

type Props = {
  skills: Skill[];
};

export default function TriggerExecutionForm({ skills }: Props) {
  const t = useTranslations("executions");
  const tc = useTranslations("common");
  const toast = useToast();
  const [isOpen, setIsOpen] = useState(false);
  const [skillId, setSkillId] = useState("");
  const [triggeredBy, setTriggeredBy] = useState("");
  const [input, setInput] = useState("");

  const createExecutionMutation = useApiMutation(createExecution, {
    queryKeysToInvalidate: [["executions"]],
    successMessage: tc("success") + ": Execution triggered",
    onSuccess: () => {
      setSkillId("");
      setTriggeredBy("");
      setInput("");
      setIsOpen(false);
    },
  });

  function handleSubmit(e: FormEvent) {
    e.preventDefault();

    let parsedInput: Record<string, unknown> = {};
    if (input.trim()) {
      try {
        parsedInput = JSON.parse(input);
      } catch {
        toast.error(tc("errors.inputInvalid"));
        return;
      }
    }

    createExecutionMutation.mutate({
      skill_id: skillId,
      triggered_by: triggeredBy,
      input: parsedInput,
    });
  }

  return (
    <>
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="trigger-btn"
      >
        {isOpen ? tc("cancel") : t("triggerExecution")}
      </button>

      {isOpen && (
        <form onSubmit={handleSubmit} className="form-panel">
          <div className="panel-header">
            <p className="section-kicker">{t("eyebrow")}</p>
            <h2>{t("trigger")}</h2>
          </div>

          {createExecutionMutation.isError && (
            <p className="form-error">{createExecutionMutation.error?.message || "Failed to trigger execution"}</p>
          )}

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
              {t("triggeredBy")}
              <input
                type="text"
                value={triggeredBy}
                onChange={(e) => setTriggeredBy(e.target.value)}
                placeholder={t("triggeredByPlaceholder")}
                required
                className="form-input"
              />
            </label>

            <label className="form-label">
              {t("input")}
              <textarea
                value={input}
                onChange={(e) => setInput(e.target.value)}
                placeholder={t("inputPlaceholder")}
                rows={3}
                className="form-textarea"
              />
            </label>
          </div>

          <button type="submit" disabled={createExecutionMutation.isPending} className="form-submit">
            {createExecutionMutation.isPending ? t("triggering") : t("trigger")}
          </button>
        </form>
      )}
    </>
  );
}
