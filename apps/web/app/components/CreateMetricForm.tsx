"use client";

import { useState, FormEvent } from "react";
import { useRouter } from "next/navigation";
import { createMetric } from "../lib/api";
import { useTranslations } from "next-intl";

type Props = {
  onCreated?: () => void;
};

export default function CreateMetricForm({ onCreated }: Props) {
  const t = useTranslations("evaluations");
  const te = useTranslations("errors");
  const router = useRouter();
  const [name, setName] = useState("");
  const [type, setType] = useState("exact_match");
  const [config, setConfig] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError("");

    let parsedConfig: Record<string, unknown> = {};
    if (config.trim()) {
      try {
        parsedConfig = JSON.parse(config);
      } catch {
        setError(te("configInvalid"));
        setLoading(false);
        return;
      }
    }

    try {
      await createMetric({
        name,
        type,
        config: parsedConfig,
      });
      setName("");
      setType("exact_match");
      setConfig("");
      onCreated?.();
      router.refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create metric");
    } finally {
      setLoading(false);
    }
  }

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

      <button type="submit" disabled={loading} className="form-submit">
        {loading ? t("creating") : t("createMetric")}
      </button>
    </form>
  );
}