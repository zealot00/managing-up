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
      {createDatasetMutation.error && <p className="form-error">{createDatasetMutation.error.message}</p>}

      <div className="form-fields">
        <label className="form-label" htmlFor="name">
          <span className="flex items-center gap-1">
            {t("datasetName")}
            <span className="text-red-500 text-sm" aria-hidden="true">*</span>
            <span className="sr-only">(required)</span>
          </span>
          <input
            type="text"
            id="name"
            {...register("name")}
            placeholder={t("datasetNamePlaceholder")}
            className={`form-input ${errors.name ? "border-red-500" : ""}`}
            aria-required="true"
          />
          {errors.name && <p className="form-error">{errors.name.message}</p>}
        </label>

        <label className="form-label" htmlFor="version">
          <span className="flex items-center gap-1">
            {t("datasetVersion")}
            <span className="text-red-500 text-sm" aria-hidden="true">*</span>
            <span className="sr-only">(required)</span>
          </span>
          <input
            type="text"
            id="version"
            {...register("version")}
            placeholder={t("datasetVersionPlaceholder")}
            className={`form-input ${errors.version ? "border-red-500" : ""}`}
            aria-required="true"
          />
          {errors.version && <p className="form-error">{errors.version.message}</p>}
        </label>

        <label className="form-label" htmlFor="owner">
          <span className="flex items-center gap-1">
            {t("datasetOwner")}
            <span className="text-red-500 text-sm" aria-hidden="true">*</span>
            <span className="sr-only">(required)</span>
          </span>
          <input
            type="text"
            id="owner"
            {...register("owner")}
            placeholder={t("datasetOwnerPlaceholder")}
            className={`form-input ${errors.owner ? "border-red-500" : ""}`}
            aria-required="true"
          />
          {errors.owner && <p className="form-error">{errors.owner.message}</p>}
        </label>

        <label className="form-label" htmlFor="description">
          {t("datasetDescription")}
          <textarea
            id="description"
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