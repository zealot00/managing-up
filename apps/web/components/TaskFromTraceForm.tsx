"use client";

import { useState } from "react";
import { buildTaskFromTrace, type Task } from "../app/lib/api";
import { useTranslations } from "next-intl";
import { useToast } from "./ToastProvider";

interface TaskFromTraceFormProps {
  onTaskCreated?: (task: Task) => void;
}

export default function TaskFromTraceForm({ onTaskCreated }: TaskFromTraceFormProps) {
  const t = useTranslations("tasks");
  const te = useTranslations("errors");
  const tc = useTranslations("common");
  const toast = useToast();
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
      setError(te("executionOrTraceRequired"));
      return;
    }

    setLoading(true);
    setError(null);
    setCreatedTask(null);

    try {
      const task = await buildTaskFromTrace(form);
      setCreatedTask(task);
      toast.success(tc("success") + ": Task built from trace");
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
          <p className="section-kicker">{t("taskBuilder")}</p>
          <h2>{t("taskBuilder.title")}</h2>
        </div>

        {error && <p className="form-error">{error}</p>}

        <div className="form-fields">
          <label className="form-label">
            {t("taskBuilder.executionId")}
            <input
              id="execution_id"
              name="execution_id"
              type="text"
              placeholder={t("taskBuilder.executionIdPlaceholder")}
              value={form.execution_id}
              onChange={(e) => setForm({ ...form, execution_id: e.target.value })}
              disabled={loading}
              className="form-input"
            />
            <span className="form-hint">
              {t("taskBuilder.executionIdHint")}
            </span>
          </label>

          <label className="form-label">
            {t("taskBuilder.traceId")}
            <input
              id="trace_id"
              name="trace_id"
              type="text"
              placeholder={t("taskBuilder.traceIdPlaceholder")}
              value={form.trace_id}
              onChange={(e) => setForm({ ...form, trace_id: e.target.value })}
              disabled={loading}
              className="form-input"
            />
            <span className="form-hint">{t("taskBuilder.traceIdHint")}</span>
          </label>
        </div>

        <button type="submit" disabled={loading} className="form-submit">
          {loading ? t("taskBuilder.building") : t("taskBuilder.buildTask")}
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
  const t = useTranslations("tasks");
  const tc = useTranslations("common");

  return (
    <article className="form-panel">
      <div className="panel-header">
        <p className="section-kicker">{t("taskBuilder.preview")}</p>
        <h2>{t("taskBuilder.generatedTask")}</h2>
        <span className="badge badge-succeeded" style={{ marginLeft: "var(--space-3)" }}>{tc("success")}</span>
      </div>

      <div className="detail-grid">
        <div className="detail-row">
          <span className="detail-label">{tc("id")}</span>
          <span className="detail-value"><code>{task.id}</code></span>
        </div>
        <div className="detail-row">
          <span className="detail-label">{tc("name")}</span>
          <span className="detail-value">{task.name}</span>
        </div>
        {task.description && (
          <div className="detail-row">
            <span className="detail-label">{tc("description")}</span>
            <span className="detail-value">{task.description}</span>
          </div>
        )}
        <div className="detail-row">
          <span className="detail-label">{t("difficulty")}</span>
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
          <h3 className="section-kicker" style={{ marginBottom: "var(--space-2)" }}>{t("taskBuilder.extractedTestCases")}</h3>
          <pre className="json-block">
            {JSON.stringify(task.test_cases, null, 2)}
          </pre>
        </div>
      )}
    </article>
  );
}