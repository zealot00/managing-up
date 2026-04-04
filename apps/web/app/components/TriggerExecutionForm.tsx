"use client";

import { useState, FormEvent } from "react";
import { useRouter } from "next/navigation";
import { createExecution, Skill } from "../lib/api";
import { useTranslations } from "next-intl";
import { useToast } from "../../components/ToastProvider";

type Props = {
  skills: Skill[];
};

export default function TriggerExecutionForm({ skills }: Props) {
  const t = useTranslations("executions");
  const tc = useTranslations("common");
  const router = useRouter();
  const toast = useToast();
  const [isOpen, setIsOpen] = useState(false);
  const [skillId, setSkillId] = useState("");
  const [triggeredBy, setTriggeredBy] = useState("");
  const [input, setInput] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError("");

    let parsedInput: Record<string, unknown> = {};
    if (input.trim()) {
      try {
        parsedInput = JSON.parse(input);
      } catch {
        setError(tc("errors.inputInvalid"));
        setLoading(false);
        return;
      }
    }

    try {
      await createExecution({
        skill_id: skillId,
        triggered_by: triggeredBy,
        input: parsedInput,
      });
      setSkillId("");
      setTriggeredBy("");
      setInput("");
      setIsOpen(false);
      toast.success(tc("success") + ": Execution triggered");
      router.refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to trigger execution");
    } finally {
      setLoading(false);
    }
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

          {error && <p className="form-error">{error}</p>}

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

          <button type="submit" disabled={loading} className="form-submit">
            {loading ? t("triggering") : t("trigger")}
          </button>
        </form>
      )}
    </>
  );
}
