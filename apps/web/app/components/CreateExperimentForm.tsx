"use client";

import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { createExperiment, Task } from "../lib/api";
import { useTranslations } from "next-intl";
import { useApiMutation } from "../lib/use-mutations";
import { createExperimentSchema } from "../lib/form-schemas";

type Props = {
  tasks: Task[];
  onCreated?: () => void;
};

export default function CreateExperimentForm({ tasks, onCreated }: Props) {
  const t = useTranslations("experiments");
  const tc = useTranslations("common");
  const createExperimentMutation = useApiMutation(createExperiment, {
    queryKeysToInvalidate: [["experiments"]],
    successMessage: tc("success") + ": Experiment created",
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
    resolver: zodResolver(createExperimentSchema),
  });

  function onSubmit(data: z.infer<typeof createExperimentSchema>) {
    createExperimentMutation.mutate({
      name: data.name,
      description: data.description ?? "",
      task_ids: data.task_ids?.split(",").map((t) => t.trim()).filter(Boolean) ?? [],
      agent_ids: data.agent_ids?.split(",").map((a) => a.trim()).filter(Boolean) ?? [],
    });
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="form-panel">
      <div className="panel-header">
        <p className="section-kicker">{t("eyebrow")}</p>
        <h2>{t("createExperiment")}</h2>
      </div>

      {createExperimentMutation.error && <p className="form-error">{createExperimentMutation.error.message}</p>}

      <div className="form-fields">
        <label className="form-label">
          {t("experimentName")}
          <input
            type="text"
            {...register("name")}
            placeholder={t("experimentNamePlaceholder")}
            className={`form-input ${errors.name ? "border-red-500" : ""}`}
          />
          {errors.name && <p className="form-error">{errors.name.message}</p>}
        </label>

        <label className="form-label">
          {tc("description")}
          <textarea
            {...register("description")}
            placeholder="What are you testing?"
            rows={2}
            className="form-textarea"
          />
        </label>

        <label className="form-label">
          {t("taskIds")}
          <input
            type="text"
            {...register("task_ids")}
            placeholder={t("taskIdsPlaceholder")}
            className="form-input"
          />
        </label>

        <label className="form-label">
          {t("agentIds")}
          <input
            type="text"
            {...register("agent_ids")}
            placeholder={t("agentIdsPlaceholder")}
            className="form-input"
          />
        </label>
      </div>

      <button type="submit" disabled={isSubmitting} className="form-submit">
        {isSubmitting ? t("creating") : t("createExperiment")}
      </button>
    </form>
  );
}