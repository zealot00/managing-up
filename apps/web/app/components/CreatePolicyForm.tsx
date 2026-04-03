"use client";

import { useState, FormEvent } from "react";
import { createSEHPolicy } from "../lib/seh-api";

type Props = {
  onCreated?: () => void;
};

export default function CreatePolicyForm({ onCreated }: Props) {
  const [name, setName] = useState("");
  const [requireProvenance, setRequireProvenance] = useState(false);
  const [requireApprovedForScore, setRequireApprovedForScore] = useState(false);
  const [minSourceDiversity, setMinSourceDiversity] = useState(2);
  const [minGoldenWeight, setMinGoldenWeight] = useState(0.5);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError("");

    try {
      await createSEHPolicy({
        name,
        require_provenance: requireProvenance,
        require_approved_for_score: requireApprovedForScore,
        min_source_diversity: minSourceDiversity,
        min_golden_weight: minGoldenWeight,
      });
      setName("");
      setRequireProvenance(false);
      setRequireApprovedForScore(false);
      setMinSourceDiversity(2);
      setMinGoldenWeight(0.5);
      onCreated?.();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create policy");
    } finally {
      setLoading(false);
    }
  }

  return (
    <form onSubmit={handleSubmit} className="form-panel">
      <div className="panel-header">
        <p className="section-kicker">SEH Module</p>
        <h2>Create governance policy</h2>
      </div>

      {error && <p className="form-error">{error}</p>}

      <div className="form-fields">
        <label className="form-label">
          Policy name
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="e.g. strict_release_policy"
            required
            className="form-input"
          />
        </label>

        <div className="checkbox-group">
          <label className="checkbox-label">
            <input
              type="checkbox"
              checked={requireProvenance}
              onChange={(e) => setRequireProvenance(e.target.checked)}
            />
            Require provenance
          </label>

          <label className="checkbox-label">
            <input
              type="checkbox"
              checked={requireApprovedForScore}
              onChange={(e) => setRequireApprovedForScore(e.target.checked)}
            />
            Require approved for score
          </label>
        </div>

        <div className="form-row">
          <label className="form-label">
            Min source diversity
            <input
              type="number"
              value={minSourceDiversity}
              onChange={(e) => setMinSourceDiversity(Number(e.target.value))}
              min={1}
              className="form-input"
            />
          </label>

          <label className="form-label">
            Min golden weight
            <input
              type="number"
              value={minGoldenWeight}
              onChange={(e) => setMinGoldenWeight(Number(e.target.value))}
              min={0}
              max={1}
              step={0.1}
              className="form-input"
            />
          </label>
        </div>
      </div>

      <button type="submit" disabled={loading} className="form-submit">
        {loading ? "Creating..." : "Create policy"}
      </button>
    </form>
  );
}
