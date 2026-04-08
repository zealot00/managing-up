"use client";

import { useState } from "react";
import { buildTaskFromTrace, type Task } from "../app/lib/api";
import { useTranslations } from "next-intl";
import { useToast } from "./ToastProvider";
import { Badge } from "../app/components/ui/Badge";
import { CheckCircle, ArrowRight, Sparkles } from "lucide-react";

interface TaskFromTraceFormProps {
  onTaskCreated?: (task: Task) => void;
  initialExecutionId?: string;
  initialTraceId?: string;
}

type ImportMode = "execution" | "trace";

export default function TaskFromTraceForm({ onTaskCreated, initialExecutionId, initialTraceId }: TaskFromTraceFormProps) {
  const t = useTranslations("tasks");
  const te = useTranslations("errors");
  const tc = useTranslations("common");
  const toast = useToast();
  const [importMode, setImportMode] = useState<ImportMode>(initialExecutionId ? "execution" : "trace");
  const [executionId, setExecutionId] = useState(initialExecutionId || "");
  const [traceId, setTraceId] = useState(initialTraceId || "");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [createdTask, setCreatedTask] = useState<Task | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const id = importMode === "execution" ? executionId : traceId;
    if (!id.trim()) {
      setError(te("executionOrTraceRequired"));
      return;
    }

    setLoading(true);
    setError(null);
    setCreatedTask(null);

    try {
      const task = await buildTaskFromTrace({
        execution_id: importMode === "execution" ? id : undefined,
        trace_id: importMode === "trace" ? id : undefined,
      });
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
          <p className="section-kicker">{t("taskBuilder.eyebrow")}</p>
          <h2>{t("taskBuilder.title")}</h2>
        </div>

        {error && <p className="form-error">{error}</p>}

        <div style={{ marginBottom: "var(--space-5)" }}>
          <p style={{ fontSize: "var(--text-sm)", color: "var(--muted)", marginBottom: "var(--space-3)" }}>
            {t("taskBuilder.importModeHint")}
          </p>
          <div className="tabs" style={{ display: "flex", gap: "var(--space-2)", borderBottom: "none" }}>
            <button
              type="button"
              onClick={() => { setImportMode("execution"); setExecutionId(""); setTraceId(""); }}
              className={`btn ${importMode === "execution" ? "btn-primary" : "btn-secondary"}`}
              style={{ flex: 1 }}
            >
              {t("taskBuilder.byExecutionId")}
            </button>
            <button
              type="button"
              onClick={() => { setImportMode("trace"); setExecutionId(""); setTraceId(""); }}
              className={`btn ${importMode === "trace" ? "btn-primary" : "btn-secondary"}`}
              style={{ flex: 1 }}
            >
              {t("taskBuilder.byTraceId")}
            </button>
          </div>
        </div>

        <div className="form-fields">
          {importMode === "execution" ? (
            <label className="form-label">
              {t("taskBuilder.executionId")}
              <input
                id="execution_id"
                name="execution_id"
                type="text"
                placeholder={t("taskBuilder.executionIdPlaceholder")}
                value={executionId}
                onChange={(e) => setExecutionId(e.target.value)}
                disabled={loading}
                className="form-input"
                autoFocus
              />
              <span className="form-hint">{t("taskBuilder.executionIdHint")}</span>
            </label>
          ) : (
            <label className="form-label">
              {t("taskBuilder.traceId")}
              <input
                id="trace_id"
                name="trace_id"
                type="text"
                placeholder={t("taskBuilder.traceIdPlaceholder")}
                value={traceId}
                onChange={(e) => setTraceId(e.target.value)}
                disabled={loading}
                className="form-input"
                autoFocus
              />
              <span className="form-hint">{t("taskBuilder.traceIdHint")}</span>
            </label>
          )}
        </div>

        <button type="submit" disabled={loading} className="form-submit">
          {loading ? (
            <>
              <Sparkles size={16} className="animate-spin" style={{ animation: "spin 1s linear infinite" }} />
              {t("taskBuilder.building")}
            </>
          ) : (
            <>
              <Sparkles size={16} />
              {t("taskBuilder.buildTask")}
            </>
          )}
        </button>
      </form>

      {createdTask && <TaskPreview task={createdTask} />}
    </>
  );
}

function TaskPreview({ task }: { task: Task }) {
  const t = useTranslations("tasks");
  const tc = useTranslations("common");

  return (
    <article className="form-panel" style={{ marginTop: "var(--space-6)", border: "1px solid var(--success)", background: "var(--success-bg)" }}>
      <div className="panel-header" style={{ marginBottom: "var(--space-4)" }}>
        <div style={{ display: "flex", alignItems: "center", gap: "var(--space-3)" }}>
          <CheckCircle size={20} style={{ color: "var(--success)" }} />
          <p className="section-kicker" style={{ margin: 0 }}>{t("taskBuilder.preview")}</p>
        </div>
        <h2>{t("taskBuilder.generatedTask")}</h2>
      </div>

      <div className="detail-grid">
        <div className="detail-row">
          <span className="detail-label">{tc("id")}</span>
          <span className="detail-value"><code style={{ fontSize: "var(--text-xs)", background: "var(--bg)", padding: "2px 6px", borderRadius: "var(--radius-xs)" }}>{task.id}</code></span>
        </div>
        <div className="detail-row">
          <span className="detail-label">{tc("name")}</span>
          <span className="detail-value" style={{ fontWeight: 600 }}>{task.name}</span>
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
            <Badge variant={task.difficulty}>{task.difficulty}</Badge>
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
          <h3 className="section-kicker" style={{ marginBottom: "var(--space-3)" }}>
            {t("taskBuilder.extractedTestCases")} ({task.test_cases.length})
          </h3>
          <div className="table-wrapper" style={{ border: "1px solid var(--line)", borderRadius: "var(--radius-md)", overflow: "hidden" }}>
            <table className="table" style={{ margin: 0 }}>
              <thead>
                <tr>
                  <th style={{ width: 40 }}>#</th>
                  <th>Input / Prompt</th>
                  <th>Expected Output</th>
                </tr>
              </thead>
              <tbody>
                {task.test_cases.map((tc, index) => (
                  <tr key={index}>
                    <td style={{ textAlign: "center", color: "var(--muted)", fontWeight: 600 }}>{index + 1}</td>
                    <td>
                      <div style={{
                        maxWidth: 300,
                        maxHeight: 100,
                        overflow: "hidden",
                        textOverflow: "ellipsis",
                        whiteSpace: "pre-wrap",
                        fontSize: "var(--text-sm)",
                        fontFamily: "monospace",
                        background: "var(--bg)",
                        padding: "var(--space-2)",
                        borderRadius: "var(--radius-sm)",
                      }}>
                        {typeof tc.input === 'string' ? tc.input : JSON.stringify(tc.input, null, 2)}
                      </div>
                    </td>
                    <td>
                      <div style={{
                        maxWidth: 300,
                        maxHeight: 100,
                        overflow: "hidden",
                        textOverflow: "ellipsis",
                        whiteSpace: "pre-wrap",
                        fontSize: "var(--text-sm)",
                        fontFamily: "monospace",
                        background: "var(--bg)",
                        padding: "var(--space-2)",
                        borderRadius: "var(--radius-sm)",
                      }}>
                        {typeof tc.expected === 'string' ? tc.expected : JSON.stringify(tc.expected, null, 2)}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </article>
  );
}
