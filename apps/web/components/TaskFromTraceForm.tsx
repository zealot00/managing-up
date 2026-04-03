"use client";

import { useState } from "react";
import { buildTaskFromTrace, type Task } from "../app/lib/api";

interface TaskFromTraceFormProps {
  onTaskCreated?: (task: Task) => void;
}

export default function TaskFromTraceForm({ onTaskCreated }: TaskFromTraceFormProps) {
  const [form, setForm] = useState({
    execution_id: "",
    trace_id: "",
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [createdTask, setCreatedTask] = useState<Task | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!form.execution_id && !form.trace_id) {
      setError("Either execution_id or trace_id is required");
      return;
    }

    setLoading(true);
    setError(null);
    setCreatedTask(null);

    try {
      const task = await buildTaskFromTrace(form);
      setCreatedTask(task);
      onTaskCreated?.(task);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to build task from trace");
    } finally {
      setLoading(false);
    }
  };

  return (
    <>
      <form onSubmit={handleSubmit} className="form-panel">
        <div className="panel-header">
          <p className="section-kicker">Task Builder</p>
          <h2>Build task from trace</h2>
        </div>

        {error && <p className="form-error">{error}</p>}

        <div className="form-fields">
          <label className="form-label">
            Execution ID
            <input
              id="execution_id"
              name="execution_id"
              type="text"
              placeholder="e.g., exec_001"
              value={form.execution_id}
              onChange={(e) => setForm({ ...form, execution_id: e.target.value })}
              disabled={loading}
              className="form-input"
            />
            <span className="form-hint">
              The execution to extract the task from (mutually exclusive with trace_id)
            </span>
          </label>

          <label className="form-label">
            Trace ID
            <input
              id="trace_id"
              name="trace_id"
              type="text"
              placeholder="e.g., trace_001"
              value={form.trace_id}
              onChange={(e) => setForm({ ...form, trace_id: e.target.value })}
              disabled={loading}
              className="form-input"
            />
            <span className="form-hint">Alternative to execution_id for trace-based extraction</span>
          </label>
        </div>

        <button type="submit" disabled={loading} className="form-submit">
          {loading ? "Building..." : "Build task from trace"}
        </button>
      </form>

      {createdTask && (
        <div style={{ marginTop: "var(--space-6)" }}>
          <TaskPreview task={createdTask} />
        </div>
      )}
    </>
  );
}

function TaskPreview({ task }: { task: Task }) {
  return (
    <article className="form-panel">
      <div className="panel-header">
        <p className="section-kicker">Preview</p>
        <h2>Generated Task</h2>
        <span className="badge badge-succeeded" style={{ marginLeft: "var(--space-3)" }}>Success</span>
      </div>

      <div className="detail-grid">
        <div className="detail-row">
          <span className="detail-label">ID</span>
          <span className="detail-value"><code>{task.id}</code></span>
        </div>
        <div className="detail-row">
          <span className="detail-label">Name</span>
          <span className="detail-value">{task.name}</span>
        </div>
        {task.description && (
          <div className="detail-row">
            <span className="detail-label">Description</span>
            <span className="detail-value">{task.description}</span>
          </div>
        )}
        <div className="detail-row">
          <span className="detail-label">Difficulty</span>
          <span className="detail-value">
            <span className={`badge badge-${task.difficulty === "easy" ? "succeeded" : task.difficulty === "medium" ? "running" : "failed"}`}>
              {task.difficulty}
            </span>
          </span>
        </div>
      </div>

      {task.tags && task.tags.length > 0 && (
        <div className="tags" style={{ marginTop: "var(--space-4)" }}>
          {task.tags.map((tag) => (
            <span key={tag} className="tag">{tag}</span>
          ))}
        </div>
      )}

      {task.test_cases && task.test_cases.length > 0 && (
        <div style={{ marginTop: "var(--space-5)" }}>
          <h3 className="section-kicker" style={{ marginBottom: "var(--space-2)" }}>Extracted Test Cases</h3>
          <pre className="json-block">
            {JSON.stringify(task.test_cases, null, 2)}
          </pre>
        </div>
      )}
    </article>
  );
}
