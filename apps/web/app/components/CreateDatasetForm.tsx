"use client";

import { useState, FormEvent } from "react";
import { createSEHDataset } from "../lib/seh-api";
import { useTranslations } from "next-intl";
import { useApiMutation } from "../lib/use-mutations";

type Props = {
  onCreated?: () => void;
};

export default function CreateDatasetForm({ onCreated }: Props) {
  const t = useTranslations("seh");
  const tc = useTranslations("common");
  const [name, setName] = useState("");
  const [version, setVersion] = useState("");
  const [owner, setOwner] = useState("");
  const [description, setDescription] = useState("");

  const createDatasetMutation = useApiMutation(createSEHDataset, {
    queryKeysToInvalidate: [["datasets"]],
    successMessage: tc("success") + ": Dataset created",
    onSuccess: () => {
      setName("");
      setVersion("");
      setOwner("");
      setDescription("");
      onCreated?.();
    },
  });

  function handleSubmit(e: FormEvent) {
    e.preventDefault();
    createDatasetMutation.mutate({ name, version, owner, description });
  }

  return (
    <form onSubmit={handleSubmit} className="form-panel">
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
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder={t("datasetNamePlaceholder")}
            required
            className="form-input"
          />
        </label>

        <label className="form-label">
          {t("datasetVersion")}
          <input
            type="text"
            value={version}
            onChange={(e) => setVersion(e.target.value)}
            placeholder={t("datasetVersionPlaceholder")}
            required
            className="form-input"
          />
        </label>

        <label className="form-label">
          {t("datasetOwner")}
          <input
            type="text"
            value={owner}
            onChange={(e) => setOwner(e.target.value)}
            placeholder={t("datasetOwnerPlaceholder")}
            required
            className="form-input"
          />
        </label>

        <label className="form-label">
          {t("datasetDescription")}
          <textarea
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder={t("datasetDescriptionPlaceholder")}
            rows={2}
            className="form-textarea"
          />
        </label>
      </div>

      <button type="submit" disabled={createDatasetMutation.isPending} className="form-submit">
        {createDatasetMutation.isPending ? t("creating") : t("createDataset")}
      </button>
    </form>
  );
}