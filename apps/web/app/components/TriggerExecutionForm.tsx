"use client";

import { useState, FormEvent } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { createExecution, Skill } from "../lib/api";
import { useApiMutation } from "../lib/use-mutations";
import { useTranslations } from "next-intl";
import { triggerExecutionSchema } from "../lib/form-schemas";
import { useToast } from "../../components/ToastProvider";

type Props = {
  skills: Skill[];
};

export default function TriggerExecutionForm({ skills }: Props) {
  const t = useTranslations("executions");
  const tc = useTranslations("common");
  const toast = useToast();
  const [isOpen, setIsOpen] = useState(false);

  const createExecutionMutation = useApiMutation(createExecution, {
    queryKeysToInvalidate: [["executions"]],
    successMessage: tc("success") + ": Execution triggered",
    onSuccess: () => {
      reset();
      setIsOpen(false);
    },
  });

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting },
  } = useForm({
    resolver: zodResolver(triggerExecutionSchema),
    defaultValues: {
      skill_id: "",
      triggered_by: "",
      input: "",
    },
  });

  function onSubmit(data: z.infer<typeof triggerExecutionSchema>) {
    let parsedInput: Record<string, unknown> = {};
    if (data.input?.trim()) {
      try {
        parsedInput = JSON.parse(data.input);
      } catch {
        toast.error(tc("errors.inputInvalid"));
        return;
      }
    }

    createExecutionMutation.mutate({
      skill_id: data.skill_id,
      triggered_by: data.triggered_by,
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
        <form onSubmit={handleSubmit(onSubmit)} className="form-panel">
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
                {...register("skill_id")}
                className={`form-select ${errors.skill_id ? "border-red-500" : ""}`}
              >
                <option value="">{t("selectSkill")}</option>
                {skills.map((s) => (
                  <option key={s.id} value={s.id}>
                    {s.name} ({s.owner_team})
                  </option>
                ))}
              </select>
              {errors.skill_id && <p className="form-error">{errors.skill_id.message}</p>}
            </label>

            <label className="form-label">
              {t("triggeredBy")}
              <input
                type="text"
                {...register("triggered_by")}
                placeholder={t("triggeredByPlaceholder")}
                className={`form-input ${errors.triggered_by ? "border-red-500" : ""}`}
              />
              {errors.triggered_by && <p className="form-error">{errors.triggered_by.message}</p>}
            </label>

            <label className="form-label">
              {t("input")}
              <textarea
                {...register("input")}
                placeholder={t("inputPlaceholder")}
                rows={3}
                className="form-textarea"
              />
              {errors.input && <p className="form-error">{errors.input.message}</p>}
            </label>
          </div>

          <button type="submit" disabled={isSubmitting} className="form-submit">
            {isSubmitting ? t("triggering") : t("trigger")}
          </button>
        </form>
      )}
    </>
  );
}