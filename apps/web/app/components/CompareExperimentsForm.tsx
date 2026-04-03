"use client";

import { useState, FormEvent } from "react";
import { compareExperiments, Experiment } from "../lib/api";

type Props = {
  experiments: Experiment[];
};

export default function CompareExperimentsForm({ experiments }: Props) {
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
      setError("Select two experiments to compare");
      return;
    }
    if (expA === expB) {
      setError("Select two different experiments");
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
        <p className="section-kicker">Experiment DB</p>
        <h2>Compare experiments</h2>
      </div>

      {error && <p className="form-error">{error}</p>}

      <form onSubmit={handleSubmit}>
        <div className="form-fields">
          <div className="form-row">
            <label className="form-label">
              Experiment A
              <select
                value={expA}
                onChange={(e) => setExpA(e.target.value)}
                className="form-select"
              >
                <option value="">Select...</option>
                {experiments.map((e) => (
                  <option key={e.id} value={e.id}>
                    {e.name}
                  </option>
                ))}
              </select>
            </label>

            <label className="form-label">
              Experiment B
              <select
                value={expB}
                onChange={(e) => setExpB(e.target.value)}
                className="form-select"
              >
                <option value="">Select...</option>
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
          {loading ? "Comparing..." : "Compare"}
        </button>
      </form>

      {result && (
        <div style={{ marginTop: "var(--space-6)" }}>
          <div className="detail-grid">
            <div className="detail-row">
              <span className="detail-label">Experiment A</span>
              <span className="detail-value">{result.experiment}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">Experiment B</span>
              <span className="detail-value">{result.compare_with}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">Regression</span>
              <span className="detail-value">
                <span className={`badge badge-${result.regression ? "failed" : "succeeded"}`}>
                  {result.regression ? "Detected" : "None"}
                </span>
              </span>
            </div>
          </div>

          {result.deltas.length > 0 && (
            <div className="gateway-table-wrapper" style={{ marginTop: "var(--space-4)" }}>
              <table className="gateway-table">
                <thead>
                  <tr>
                    <th>Task ID</th>
                    <th>A Score</th>
                    <th>B Score</th>
                    <th>Delta</th>
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
