"use client";

import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { createSEHDataset } from "../lib/seh-api";
import { useTranslations } from "next-intl";
import { useApiMutation } from "../lib/use-mutations";
import { createDatasetSchema } from "../lib/form-schemas";

type Props = {
  onCreated?: () => void;
};

export default function CreateDatasetForm({ onCreated }: Props) {
  const t = useTranslations("seh");
  const tc = useTranslations("common");
  const createDatasetMutation = useApiMutation(createSEHDataset, {
    queryKeysToInvalidate: [["datasets"]],
    successMessage: tc("success") + ": Dataset created",
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
    resolver: zodResolver(createDatasetSchema),
  });

  function onSubmit(data: z.infer<typeof createDatasetSchema>) {
    createDatasetMutation.mutate({
      name: data.name,
      version: data.version,
      owner: data.owner,
      description: data.description,
    });
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="form-panel">
      <div className="panel-header">
        <p className="section-kicker">{t("eyebrow")}</p>
        <h2>{t("createDataset")}</h2>
      </div>

      {createDatasetMutation.error && <p className="form-error">{createDatasetMutation.error.message}</p>}

      <div className="form-fields">
        <label className="form-label">
          {t("datasetName")}
          <input
            type="text"
            {...register("name")}
            placeholder={t("datasetNamePlaceholder")}
            className={`form-input ${errors.name ? "border-red-500" : ""}`}
          />
          {errors.name && <p className="form-error">{errors.name.message}</p>}
        </label>

        <label className="form-label">
          {t("datasetVersion")}
          <input
            type="text"
            {...register("version")}
            placeholder={t("datasetVersionPlaceholder")}
            className={`form-input ${errors.version ? "border-red-500" : ""}`}
          />
          {errors.version && <p className="form-error">{errors.version.message}</p>}
        </label>

        <label className="form-label">
          {t("datasetOwner")}
          <input
            type="text"
            {...register("owner")}
            placeholder={t("datasetOwnerPlaceholder")}
            className={`form-input ${errors.owner ? "border-red-500" : ""}`}
          />
          {errors.owner && <p className="form-error">{errors.owner.message}</p>}
        </label>

        <label className="form-label">
          {t("datasetDescription")}
          <textarea
            {...register("description")}
            placeholder={t("datasetDescriptionPlaceholder")}
            rows={2}
            className="form-textarea"
          />
        </label>
      </div>

      <button type="submit" disabled={isSubmitting} className="form-submit">
        {isSubmitting ? t("creating") : t("createDataset")}
      </button>
    </form>
  );
}