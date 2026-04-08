"use client";

import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { updateTask, Skill, Task, CreateTaskRequest } from "../lib/api";
import { useTranslations } from "next-intl";
import { useApiMutation } from "../lib/use-mutations";
import { updateTaskSchema } from "../lib/form-schemas";
import { useToast } from "../../components/ToastProvider";

type Props = {
  task: Task;
  skills: Skill[];
  onCancel: () => void;
  onUpdated: () => void;
};

type UpdateTaskVariables = {
  id: string;
} & Partial<CreateTaskRequest>;

async function updateTaskWrapper(vars: UpdateTaskVariables) {
  const { id, ...req } = vars;
  return updateTask(id, req);
}

export default function EditTaskForm({ task, skills, onCancel, onUpdated }: Props) {
  const t = useTranslations("tasks");
  const te = useTranslations("errors");
  const toast = useToast();
  const updateMutation = useApiMutation(updateTaskWrapper, {
    successMessage: "Task updated",
    queryKeysToInvalidate: [["tasks"]],
    onSuccess: () => {
      onUpdated();
    },
  });

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm({
    resolver: zodResolver(updateTaskSchema),
    defaultValues: {
      name: task.name,
      description: task.description,
      skill_id: task.skill_id,
      difficulty: task.difficulty,
      tags: task.tags.join(", "),
      test_cases: JSON.stringify(task.test_cases, null, 2),
    },
  });

  function onSubmit(data: z.infer<typeof updateTaskSchema>) {
    let parsedTestCases: Array<{ input: Record<string, unknown>; expected: unknown }> = [];
    if (data.test_cases?.trim()) {
      try {
        parsedTestCases = JSON.parse(data.test_cases);
      } catch {
        toast.error(te("testCasesInvalid"));
        return;
      }
    }

    updateMutation.mutate({
      id: task.id,
      name: data.name,
      description: data.description,
      skill_id: data.skill_id,
      tags: data.tags?.split(",").map((t) => t.trim()).filter(Boolean),
      difficulty: data.difficulty,
      test_cases: parsedTestCases,
    });
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="form-panel">
      <div className="panel-header">
        <p className="section-kicker">{t("eyebrow")}</p>
        <h2>{t("editTask", { name: task.name })}</h2>
      </div>

      {updateMutation.isError && <p className="form-error">{updateMutation.error?.message}</p>}

      <div className="form-fields">
        <label className="form-label">
          {t("taskName")}
          <input
            type="text"
            {...register("name")}
            className={`form-input ${errors.name ? "border-red-500" : ""}`}
          />
          {errors.name && <p className="form-error">{errors.name.message}</p>}
        </label>

        <label className="form-label">
          {t("description")}
          <textarea
            {...register("description")}
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
            className="form-input"
          />
        </label>

        <label className="form-label">
          {t("testCases")}
          <textarea
            {...register("test_cases")}
            rows={3}
            className="form-textarea"
          />
          {errors.test_cases && <p className="form-error">{errors.test_cases.message}</p>}
        </label>
      </div>

      <div className="form-actions">
        <button type="submit" disabled={isSubmitting} className="form-submit" style={{ flex: 1 }}>
          {isSubmitting ? t("saving") : t("saveChanges")}
        </button>
        <button type="button" onClick={onCancel} className="btn btn-secondary" style={{ flex: 1 }}>
          {t("cancel")}
        </button>
      </div>
    </form>
  );
}