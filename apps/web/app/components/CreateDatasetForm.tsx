"use client";

import { useState, FormEvent } from "react";
import { createSEHDataset } from "../lib/seh-api";

type Props = {
  onCreated?: () => void;
};

export default function CreateDatasetForm({ onCreated }: Props) {
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
      onCreated?.();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create dataset");
    } finally {
      setLoading(false);
    }
  }

  return (
    <form onSubmit={handleSubmit} className="form-panel">
      <div className="panel-header">
        <p className="section-kicker">SEH Module</p>
        <h2>Create dataset</h2>
      </div>

      {error && <p className="form-error">{error}</p>}

      <div className="form-fields">
        <label className="form-label">
          Dataset name
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="e.g. production_errors_2026"
            required
            className="form-input"
          />
        </label>

        <label className="form-label">
          Version
          <input
            type="text"
            value={version}
            onChange={(e) => setVersion(e.target.value)}
            placeholder="e.g. 1.0.0"
            required
            className="form-input"
          />
        </label>

        <label className="form-label">
          Owner
          <input
            type="text"
            value={owner}
            onChange={(e) => setOwner(e.target.value)}
            placeholder="e.g. qa_team"
            required
            className="form-input"
          />
        </label>

        <label className="form-label">
          Description
          <textarea
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="Dataset description..."
            rows={2}
            className="form-textarea"
          />
        </label>
      </div>

      <button type="submit" disabled={loading} className="form-submit">
        {loading ? "Creating..." : "Create dataset"}
      </button>
    </form>
  );
}
