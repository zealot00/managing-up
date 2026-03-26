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
      <form onSubmit={handleSubmit}>
        {error && (
          <div className="form-error" style={{ marginBottom: 16 }}>
            {error}
          </div>
        )}

        <div className="form-group">
          <label htmlFor="execution_id">Execution ID</label>
          <input
            id="execution_id"
            name="execution_id"
            type="text"
            placeholder="e.g., exec_001"
            value={form.execution_id}
            onChange={(e) => setForm({ ...form, execution_id: e.target.value })}
            disabled={loading}
          />
          <span className="form-hint">
            The execution to extract the task from (mutually exclusive with trace_id)
          </span>
        </div>

        <div className="form-group">
          <label htmlFor="trace_id">Trace ID</label>
          <input
            id="trace_id"
            name="trace_id"
            type="text"
            placeholder="e.g., trace_001"
            value={form.trace_id}
            onChange={(e) => setForm({ ...form, trace_id: e.target.value })}
            disabled={loading}
          />
          <span className="form-hint">Alternative to execution_id for trace-based extraction</span>
        </div>

        <button type="submit" className="btn btn-primary" disabled={loading} style={{ marginTop: 16 }}>
          {loading ? "Building Task..." : "Build Task from Trace"}
        </button>
      </form>

      {createdTask && (
        <div style={{ marginTop: 32 }}>
          <div className="panel-header" style={{ marginBottom: 16 }}>
            <h2>Generated Task Preview</h2>
            <span className="badge badge-succeeded">Success</span>
          </div>

          <TaskPreview task={createdTask} />
        </div>
      )}
    </>
  );
}

function TaskPreview({ task }: { task: Task }) {
  return (
    <article className="panel">
      <div style={{ marginBottom: 16 }}>
        <strong>ID:</strong> <code>{task.id}</code>
      </div>
      <div style={{ marginBottom: 16 }}>
        <strong>Name:</strong> {task.name}
      </div>
      {task.description && (
        <div style={{ marginBottom: 16 }}>
          <strong>Description:</strong> {task.description}
        </div>
      )}
      <div style={{ marginBottom: 16 }}>
        <strong>Difficulty:</strong>{" "}
        <span
          className={`badge badge-${
            task.difficulty === "easy" ? "succeeded" : task.difficulty === "medium" ? "running" : "failed"
          }`}
        >
          {task.difficulty}
        </span>
      </div>
      {task.tags && task.tags.length > 0 && (
        <div style={{ marginBottom: 16 }}>
          <strong>Tags:</strong>
          <div style={{ display: "flex", gap: 8, marginTop: 8, flexWrap: "wrap" }}>
            {task.tags.map((tag) => (
              <span
                key={tag}
                style={{
                  padding: "4px 10px",
                  borderRadius: 999,
                  background: "rgba(36, 49, 64, 0.06)",
                  fontSize: "0.78rem",
                }}
              >
                {tag}
              </span>
            ))}
          </div>
        </div>
      )}
      {task.test_cases && task.test_cases.length > 0 && (
        <div>
          <strong>Extracted Test Cases:</strong>
          <pre
            style={{
              marginTop: 8,
              padding: 12,
              background: "rgba(36, 49, 64, 0.04)",
              borderRadius: 8,
              fontSize: "0.85rem",
              overflow: "auto",
            }}
          >
            {JSON.stringify(task.test_cases, null, 2)}
          </pre>
        </div>
      )}
    </article>
  );
}
