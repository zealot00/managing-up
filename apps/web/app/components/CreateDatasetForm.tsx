"use client";

import { useState, FormEvent } from "react";
import { useRouter } from "next/navigation";
import { createSEHDataset } from "../lib/seh-api";
import { useTranslations } from "next-intl";
import { useToast } from "../../components/ToastProvider";

type Props = {
  onCreated?: () => void;
};

export default function CreateDatasetForm({ onCreated }: Props) {
  const t = useTranslations("seh");
  const tc = useTranslations("common");
  const router = useRouter();
  const toast = useToast();
  const [name, setName] = useState("");
  const [version, setVersion] = useState("");
  const [owner, setOwner] = useState("");
  const [description, setDescription] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError("");

    try {
      await createSEHDataset({ name, version, owner, description });
      setName("");
      setVersion("");
      setOwner("");
      setDescription("");
      toast.success(tc("success") + ": Dataset created");
      onCreated?.();
      router.refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create dataset");
    } finally {
      setLoading(false);
    }
  }

  return (
    <form onSubmit={handleSubmit} className="form-panel">
      <div className="panel-header">
        <p className="section-kicker">{t("eyebrow")}</p>
        <h2>{t("createDataset")}</h2>
      </div>

      {error && <p className="form-error">{error}</p>}

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

      <button type="submit" disabled={loading} className="form-submit">
        {loading ? t("creating") : t("createDataset")}
      </button>
    </form>
  );
}