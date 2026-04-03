"use client";

import { useState, FormEvent } from "react";
import { compareExperiments, Experiment } from "../lib/api";
import { useTranslations } from "next-intl";

type Props = {
  experiments: Experiment[];
};

export default function CompareExperimentsForm({ experiments }: Props) {
  const t = useTranslations("experiments");
  const te = useTranslations("errors");
  const [expA, setExpA] = useState("");
  const [expB, setExpB] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [result, setResult] = useState<{
    experiment: string;
    compare_with: string;
    deltas: Array<{ task_id: string; exp_score: number; other_score: number; delta: number }>;
    regression: boolean;
  } | null>(null);

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    if (!expA || !expB) {
      setError(te("selectTwoExperiments"));
      return;
    }
    if (expA === expB) {
      setError(te("selectDifferentExperiments"));
      return;
    }
    setLoading(true);
    setError("");
    setResult(null);

    try {
      const data = await compareExperiments(expA, expB);
      setResult(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to compare experiments");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="form-panel">
      <div className="panel-header">
        <p className="section-kicker">{t("eyebrow")}</p>
        <h2>{t("compareTitle")}</h2>
      </div>

      {error && <p className="form-error">{error}</p>}

      <form onSubmit={handleSubmit}>
        <div className="form-fields">
          <div className="form-row">
            <label className="form-label">
              {t("experimentA")}
              <select
                value={expA}
                onChange={(e) => setExpA(e.target.value)}
                className="form-select"
              >
                <option value="">{t("select")}</option>
                {experiments.map((e) => (
                  <option key={e.id} value={e.id}>
                    {e.name}
                  </option>
                ))}
              </select>
            </label>

            <label className="form-label">
              {t("experimentB")}
              <select
                value={expB}
                onChange={(e) => setExpB(e.target.value)}
                className="form-select"
              >
                <option value="">{t("select")}</option>
                {experiments.map((e) => (
                  <option key={e.id} value={e.id}>
                    {e.name}
                  </option>
                ))}
              </select>
            </label>
          </div>
        </div>

        <button type="submit" disabled={loading} className="form-submit">
          {loading ? t("comparing") : t("compareBtn")}
        </button>
      </form>

      {result && (
        <div style={{ marginTop: "var(--space-6)" }}>
          <div className="detail-grid">
            <div className="detail-row">
              <span className="detail-label">{t("experimentA")}</span>
              <span className="detail-value">{result.experiment}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">{t("experimentB")}</span>
              <span className="detail-value">{result.compare_with}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">{t("regression")}</span>
              <span className="detail-value">
                <span className={`badge badge-${result.regression ? "failed" : "succeeded"}`}>
                  {result.regression ? t("detected") : t("none")}
                </span>
              </span>
            </div>
          </div>

          {result.deltas.length > 0 && (
            <div className="gateway-table-wrapper" style={{ marginTop: "var(--space-4)" }}>
              <table className="gateway-table">
                <thead>
                  <tr>
                    <th>{t("taskIds").split("(")[0].trim()}</th>
                    <th>{t("aScore")}</th>
                    <th>{t("bScore")}</th>
                    <th>{t("delta")}</th>
                  </tr>
                </thead>
                <tbody>
                  {result.deltas.map((d) => (
                    <tr key={d.task_id}>
                      <td>{d.task_id}</td>
                      <td>{d.exp_score.toFixed(3)}</td>
                      <td>{d.other_score.toFixed(3)}</td>
                      <td>
                        <span className={`badge badge-${d.delta >= 0 ? "succeeded" : "failed"}`}>
                          {d.delta >= 0 ? "+" : ""}{d.delta.toFixed(3)}
                        </span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}
    </div>
  );
}