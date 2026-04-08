"use client";

import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { createMetric } from "../lib/api";
import { useTranslations } from "next-intl";
import { useApiMutation } from "../lib/use-mutations";
import { createMetricSchema } from "../lib/form-schemas";
import { useToast } from "../../components/ToastProvider";

type Props = {
  onCreated?: () => void;
};

export default function CreateMetricForm({ onCreated }: Props) {
  const t = useTranslations("evaluations");
  const tc = useTranslations("common");
  const te = useTranslations("errors");
  const toast = useToast();
  const createMetricMutation = useApiMutation(createMetric, {
    queryKeysToInvalidate: [["metrics"]],
    successMessage: tc("success") + ": Metric created",
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
    resolver: zodResolver(createMetricSchema),
    defaultValues: {
      type: "exact_match",
    },
  });

  function onSubmit(data: z.infer<typeof createMetricSchema>) {
    let parsedConfig: Record<string, unknown> = {};
    if (data.config?.trim()) {
      try {
        parsedConfig = JSON.parse(data.config);
      } catch {
        toast.error(te("configInvalid"));
        return;
      }
    }

    createMetricMutation.mutate({
      name: data.name,
      type: data.type,
      config: parsedConfig,
    });
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="form-panel">
      <div className="panel-header">
        <p className="section-kicker">{t("eyebrow")}</p>
        <h2>{t("createMetric")}</h2>
      </div>

      {createMetricMutation.error && <p className="form-error">{createMetricMutation.error.message}</p>}

      <div className="form-fields">
        <label className="form-label">
          {t("metricName")}
          <input
            type="text"
            {...register("name")}
            placeholder={t("metricNamePlaceholder")}
            className={`form-input ${errors.name ? "border-red-500" : ""}`}
          />
          {errors.name && <p className="form-error">{errors.name.message}</p>}
        </label>

        <label className="form-label">
          {t("metricType")}
          <select {...register("type")} className="form-select">
            <option value="exact_match">{t("exactMatch")}</option>
            <option value="llm_judge">{t("llmJudge")}</option>
            <option value="custom">{t("custom")}</option>
          </select>
          {errors.type && <p className="form-error">{errors.type.message}</p>}
        </label>

        <label className="form-label">
          {t("config")}
          <textarea
            {...register("config")}
            placeholder={t("configPlaceholder")}
            rows={3}
            className="form-textarea"
          />
        </label>
      </div>

      <button type="submit" disabled={isSubmitting} className="form-submit">
        {isSubmitting ? t("creating") : t("createMetric")}
      </button>
    </form>
  );
}