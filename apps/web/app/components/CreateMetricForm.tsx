"use client";

import { useState, FormEvent } from "react";
import { createMetric } from "../lib/api";
import { useTranslations } from "next-intl";
import { useApiMutation } from "../lib/use-mutations";

type Props = {
  onCreated?: () => void;
};

export default function CreateMetricForm({ onCreated }: Props) {
  const t = useTranslations("evaluations");
  const tc = useTranslations("common");
  const te = useTranslations("errors");
  const [name, setName] = useState("");
  const [type, setType] = useState("exact_match");
  const [config, setConfig] = useState("");
  const [localError, setLocalError] = useState("");

  const createMetricMutation = useApiMutation(createMetric, {
    queryKeysToInvalidate: [["metrics"]],
    successMessage: tc("success") + ": Metric created",
    onSuccess: () => {
      setName("");
      setType("exact_match");
      setConfig("");
      onCreated?.();
    },
  });

  function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setLocalError("");

    let parsedConfig: Record<string, unknown> = {};
    if (config.trim()) {
      try {
        parsedConfig = JSON.parse(config);
      } catch {
        setLocalError(te("configInvalid"));
        return;
      }
    }

    createMetricMutation.mutate({
      name,
      type,
      config: parsedConfig,
    });
  }

  const error = createMetricMutation.error?.message || localError;

  return (
    <form onSubmit={handleSubmit} className="form-panel">
      <div className="panel-header">
        <p className="section-kicker">{t("eyebrow")}</p>
        <h2>{t("createMetric")}</h2>
      </div>

      {error && <p className="form-error">{error}</p>}

      <div className="form-fields">
        <label className="form-label">
          {t("metricName")}
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder={t("metricNamePlaceholder")}
            required
            className="form-input"
          />
        </label>

        <label className="form-label">
          {t("metricType")}
          <select
            value={type}
            onChange={(e) => setType(e.target.value)}
            className="form-select"
          >
            <option value="exact_match">{t("exactMatch")}</option>
            <option value="llm_judge">{t("llmJudge")}</option>
            <option value="custom">{t("custom")}</option>
          </select>
        </label>

        <label className="form-label">
          {t("config")}
          <textarea
            value={config}
            onChange={(e) => setConfig(e.target.value)}
            placeholder={t("configPlaceholder")}
            rows={3}
            className="form-textarea"
          />
        </label>
      </div>

      <button type="submit" disabled={createMetricMutation.isPending} className="form-submit">
        {createMetricMutation.isPending ? t("creating") : t("createMetric")}
      </button>
    </form>
  );
}