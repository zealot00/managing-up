"use client";

import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { useTranslations } from "next-intl";
import { useApiMutation } from "../lib/use-mutations";
import { createTask, Skill } from "../lib/api";
import { createTaskSchema } from "../lib/form-schemas";
import { useToast } from "../../components/ToastProvider";

type Props = {
  skills: Skill[];
  onCreated?: () => void;
};

export default function CreateTaskForm({ skills, onCreated }: Props) {
  const t = useTranslations("tasks");
  const te = useTranslations("errors");
  const toast = useToast();
  const createMutation = useApiMutation(createTask, {
    successMessage: "Task created",
    queryKeysToInvalidate: [["tasks"]],
    onSuccess: () => {
      reset();
      onCreated?.();
    },
  });

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting },
  } = useForm({
    resolver: zodResolver(createTaskSchema),
    defaultValues: {
      difficulty: "medium",
    },
  });

  function onSubmit(data: z.infer<typeof createTaskSchema>) {
    let parsedTestCases: Array<{ input: Record<string, unknown>; expected: unknown }> = [];
    if (data.test_cases?.trim()) {
      try {
        parsedTestCases = JSON.parse(data.test_cases);
      } catch {
        toast.error(te("testCasesInvalid"));
        return;
      }
    }

    createMutation.mutate({
      name: data.name,
      description: data.description ?? "",
      skill_id: data.skill_id ?? "",
      tags: data.tags?.split(",").map((t) => t.trim()).filter(Boolean) ?? [],
      difficulty: data.difficulty,
      test_cases: parsedTestCases,
    });
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="form-panel">
      <div className="panel-header">
        <p className="section-kicker">{t("eyebrow")}</p>
        <h2>{t("createTask")}</h2>
      </div>

      {createMutation.isError && <p className="form-error">{createMutation.error?.message}</p>}

      <div className="form-fields">
        <label className="form-label">
          {t("taskName")}
          <input
            type="text"
            {...register("name")}
            placeholder={t("taskNamePlaceholder")}
            className={`form-input ${errors.name ? "border-red-500" : ""}`}
          />
          {errors.name && <p className="form-error">{errors.name.message}</p>}
        </label>

        <label className="form-label">
          {t("description")}
          <textarea
            {...register("description")}
            placeholder="What does this task evaluate?"
            rows={2}
            className="form-textarea"
          />
        </label>

        <label className="form-label">
          {t("linkedSkill")}
          <select {...register("skill_id")} className="form-select">
            <option value="">{t("noSkill")}</option>
            {skills.map((s) => (
              <option key={s.id} value={s.id}>
                {s.name}
              </option>
            ))}
          </select>
        </label>

        <label className="form-label">
          {t("difficulty")}
          <select {...register("difficulty")} className="form-select">
            <option value="easy">{t("easy")}</option>
            <option value="medium">{t("medium")}</option>
            <option value="hard">{t("hard")}</option>
          </select>
        </label>

        <label className="form-label">
          {t("tags")}
          <input
            type="text"
            {...register("tags")}
            placeholder={t("tagsPlaceholder")}
            className="form-input"
          />
        </label>

        <label className="form-label">
          {t("testCases")}
          <textarea
            {...register("test_cases")}
            placeholder={t("testCasesPlaceholder")}
            rows={3}
            className="form-textarea"
          />
          {errors.test_cases && <p className="form-error">{errors.test_cases.message}</p>}
        </label>
      </div>

      <button type="submit" disabled={isSubmitting} className="form-submit">
        {isSubmitting ? t("creating") : t("create")}
      </button>
    </form>
  );
}