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
      {createMutation.isError && <p className="form-error">{createMutation.error?.message}</p>}

      <div className="form-fields">
        <label className="form-label" htmlFor="name">
          <span className="flex items-center gap-1">
            {t("taskName")}
            <span className="text-red-500 text-sm" aria-hidden="true">*</span>
            <span className="sr-only">(required)</span>
          </span>
          <input
            type="text"
            id="name"
            {...register("name")}
            placeholder={t("taskNamePlaceholder")}
            className={`form-input ${errors.name ? "border-red-500" : ""}`}
            aria-required="true"
          />
          {errors.name && <p className="form-error">{errors.name.message}</p>}
        </label>

        <label className="form-label" htmlFor="description">
          {t("description")}
          <textarea
            id="description"
            {...register("description")}
            placeholder="What does this task evaluate?"
            rows={2}
            className="form-textarea"
          />
        </label>

        <label className="form-label" htmlFor="skill_id">
          {t("linkedSkill")}
          <select id="skill_id" {...register("skill_id")} className="form-select">
            <option value="">{t("noSkill")}</option>
            {skills.map((s) => (
              <option key={s.id} value={s.id}>
                {s.name}
              </option>
            ))}
          </select>
        </label>

        <label className="form-label" htmlFor="difficulty">
          {t("difficulty")}
          <select id="difficulty" {...register("difficulty")} className="form-select">
            <option value="easy">{t("easy")}</option>
            <option value="medium">{t("medium")}</option>
            <option value="hard">{t("hard")}</option>
          </select>
        </label>

        <label className="form-label" htmlFor="tags">
          {t("tags")}
          <input
            type="text"
            id="tags"
            {...register("tags")}
            placeholder={t("tagsPlaceholder")}
            className="form-input"
          />
        </label>

        <label className="form-label" htmlFor="test_cases">
          {t("testCases")}
          <textarea
            id="test_cases"
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