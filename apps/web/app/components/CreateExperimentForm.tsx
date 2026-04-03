"use client";

import { useState, FormEvent } from "react";
import { useRouter } from "next/navigation";
import { createExperiment, Task } from "../lib/api";

type Props = {
  tasks: Task[];
  onCreated?: () => void;
};

export default function CreateExperimentForm({ tasks, onCreated }: Props) {
  const router = useRouter();
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [taskIds, setTaskIds] = useState("");
  const [agentIds, setAgentIds] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError("");

    try {
      await createExperiment({
        name,
        description,
        task_ids: taskIds.split(",").map((t) => t.trim()).filter(Boolean),
        agent_ids: agentIds.split(",").map((a) => a.trim()).filter(Boolean),
      });
      setName("");
      setDescription("");
      setTaskIds("");
      setAgentIds("");
      onCreated?.();
      router.refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create experiment");
    } finally {
      setLoading(false);
    }
  }

  return (
    <form onSubmit={handleSubmit} className="form-panel">
      <div className="panel-header">
        <p className="section-kicker">Experiment DB</p>
        <h2>Create experiment</h2>
      </div>

      {error && <p className="form-error">{error}</p>}

      <div className="form-fields">
        <label className="form-label">
          Experiment name
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="e.g. agent-v2-vs-v3"
            required
            className="form-input"
          />
        </label>

        <label className="form-label">
          Description
          <textarea
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="What are you testing?"
            rows={2}
            className="form-textarea"
          />
        </label>

        <label className="form-label">
          Task IDs (comma-separated)
          <input
            type="text"
            value={taskIds}
            onChange={(e) => setTaskIds(e.target.value)}
            placeholder="e.g. task_001, task_002"
            className="form-input"
          />
        </label>

        <label className="form-label">
          Agent IDs (comma-separated)
          <input
            type="text"
            value={agentIds}
            onChange={(e) => setAgentIds(e.target.value)}
            placeholder="e.g. agent-v1, agent-v2"
            className="form-input"
          />
        </label>
      </div>

      <button type="submit" disabled={loading} className="form-submit">
        {loading ? "Creating..." : "Create experiment"}
      </button>
    </form>
  );
}
