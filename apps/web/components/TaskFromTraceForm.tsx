"use client";

import { useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { buildTaskFromTrace, type Task } from "../app/lib/api";
import { useTranslations } from "next-intl";
import { useToast } from "./ToastProvider";
import { Badge } from "../app/components/ui/Badge";
import {
  CheckCircle,
  Play,
  GitBranch,
  Sparkles,
  Hash,
  FileText,
  ArrowRight,
} from "lucide-react";

interface TaskFromTraceWizardProps {
  onTaskCreated?: (task: Task) => void;
  initialExecutionId?: string;
  initialTraceId?: string;
  hideHeader?: boolean;
}

type ImportMode = "execution" | "trace";
type WizardStep = 1 | 2 | 3;

export default function TaskFromTraceWizard({
  onTaskCreated,
  initialExecutionId,
  initialTraceId,
  hideHeader,
}: TaskFromTraceWizardProps) {
  const t = useTranslations("tasks");
  const te = useTranslations("errors");
  const tc = useTranslations("common");
  const toast = useToast();
  const queryClient = useQueryClient();

  const defaultMode: ImportMode = initialExecutionId ? "execution" : initialTraceId ? "trace" : "execution";
  const [importMode, setImportMode] = useState<ImportMode>(defaultMode);
  const [executionId, setExecutionId] = useState(initialExecutionId || "");
  const [traceId, setTraceId] = useState(initialTraceId || "");
  const [step, setStep] = useState<WizardStep>(1);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [createdTask, setCreatedTask] = useState<Task | null>(null);

  const currentId = importMode === "execution" ? executionId : traceId;
  const canProceed = currentId.trim().length > 0;

  function handleNext() {
    if (step === 1 && canProceed) setStep(2);
  }

  function handleBack() {
    if (step === 2) setStep(1);
  }

  function resetWizard() {
    setStep(1);
    setExecutionId("");
    setTraceId("");
    setCreatedTask(null);
    setError(null);
  }

  async function handleBuild() {
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
      setStep(3);
      toast.success(tc("success") + ": Task built from trace");
      queryClient.invalidateQueries({ queryKey: ["tasks"] });
      onTaskCreated?.(task);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to build task from trace");
      setStep(2);
    } finally {
      setLoading(false);
    }
  }

  const steps = [
    { num: 1, label: t("taskBuilder.step1Label") },
    { num: 2, label: t("taskBuilder.step2Label") },
    { num: 3, label: t("taskBuilder.step3Label") },
  ];

  return (
    <>
      {!hideHeader && (
        <div className="wizard-steps">
          {steps.map((s, i) => (
            <span key={s.num} style={{ display: "contents" }}>
              <span
                className={`wizard-step ${
                  step === s.num
                    ? "wizard-step-active"
                    : step > s.num
                    ? "wizard-step-done"
                    : ""
                }`}
              >
                <span className="wizard-step-num">
                  {step > s.num ? "✓" : s.num}
                </span>
                {s.label}
              </span>
              {i < steps.length - 1 && <span className="wizard-step-sep" />}
            </span>
          ))}
        </div>
      )}

      {step === 1 && (
        <Step1SourceSelect
          importMode={importMode}
          executionId={executionId}
          traceId={traceId}
          onModeChange={(mode) => {
            setImportMode(mode);
            if (mode === "execution") setTraceId("");
            else setExecutionId("");
          }}
          onExecutionIdChange={setExecutionId}
          onTraceIdChange={setTraceId}
        />
      )}

      {step === 2 && (
        <Step2Confirm
          importMode={importMode}
          id={currentId}
          loading={loading}
          error={error}
          onBuild={handleBuild}
        />
      )}

      {step === 3 && createdTask && (
        <Step3Result task={createdTask} />
      )}

      <div className="wizard-actions">
        {step === 1 && (
          <>
            <span />
            <button
              className="btn btn-primary"
              disabled={!canProceed}
              onClick={handleNext}
            >
              {tc("next")} <ArrowRight size={16} />
            </button>
          </>
        )}

        {step === 2 && (
          <>
            <button className="btn btn-secondary" onClick={handleBack} disabled={loading}>
              {tc("back")}
            </button>
            <button className="btn btn-primary" onClick={handleBuild} disabled={loading}>
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
          </>
        )}

        {step === 3 && (
          <>
            <button className="btn btn-secondary" onClick={resetWizard}>
              {t("taskBuilder.continueBuild")}
            </button>
            <a href={`/tasks?highlight=${createdTask?.id}`} className="btn btn-primary">
              {t("taskBuilder.viewTask")} <ArrowRight size={16} />
            </a>
          </>
        )}
      </div>
    </>
  );
}

/* ========== Step 1: Source Select ========== */

function Step1SourceSelect({
  importMode,
  executionId,
  traceId,
  onModeChange,
  onExecutionIdChange,
  onTraceIdChange,
}: {
  importMode: ImportMode;
  executionId: string;
  traceId: string;
  onModeChange: (mode: ImportMode) => void;
  onExecutionIdChange: (val: string) => void;
  onTraceIdChange: (val: string) => void;
}) {
  const t = useTranslations("tasks");

  return (
    <div>
      <p style={{ fontSize: "var(--text-sm)", color: "var(--muted)", marginBottom: "var(--space-4)" }}>
        {t("taskBuilder.importModeHint")}
      </p>
      <div className="wizard-source-cards">
        <div
          className={`wizard-source-card ${importMode === "execution" ? "wizard-source-card-active" : ""}`}
          onClick={() => onModeChange("execution")}
          role="button"
          tabIndex={0}
          onKeyDown={(e) => e.key === "Enter" && onModeChange("execution")}
        >
          <div className="wizard-source-card-icon">
            <Play size={20} />
          </div>
          <div className="wizard-source-card-title">{t("taskBuilder.byExecutionId")}</div>
          <div className="wizard-source-card-desc">{t("taskBuilder.executionIdHint")}</div>
          {importMode === "execution" && (
            <div className="wizard-source-card-body">
              <label className="form-label" htmlFor="wizard-exec-id">
                {t("taskBuilder.executionId")}
                <input
                  id="wizard-exec-id"
                  type="text"
                  className="form-input"
                  placeholder={t("taskBuilder.executionIdPlaceholder")}
                  value={executionId}
                  onChange={(e) => onExecutionIdChange(e.target.value)}
                  autoFocus
                />
              </label>
            </div>
          )}
        </div>

        <div
          className={`wizard-source-card ${importMode === "trace" ? "wizard-source-card-active" : ""}`}
          onClick={() => onModeChange("trace")}
          role="button"
          tabIndex={0}
          onKeyDown={(e) => e.key === "Enter" && onModeChange("trace")}
        >
          <div className="wizard-source-card-icon">
            <GitBranch size={20} />
          </div>
          <div className="wizard-source-card-title">{t("taskBuilder.byTraceId")}</div>
          <div className="wizard-source-card-desc">{t("taskBuilder.traceIdHint")}</div>
          {importMode === "trace" && (
            <div className="wizard-source-card-body">
              <label className="form-label" htmlFor="wizard-trace-id">
                {t("taskBuilder.traceId")}
                <input
                  id="wizard-trace-id"
                  type="text"
                  className="form-input"
                  placeholder={t("taskBuilder.traceIdPlaceholder")}
                  value={traceId}
                  onChange={(e) => onTraceIdChange(e.target.value)}
                  autoFocus
                />
              </label>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

/* ========== Step 2: Confirm ========== */

function Step2Confirm({
  importMode,
  id,
  loading,
  error,
  onBuild,
}: {
  importMode: ImportMode;
  id: string;
  loading: boolean;
  error: string | null;
  onBuild: () => void;
}) {
  const t = useTranslations("tasks");

  return (
    <div>
      <p style={{ fontSize: "var(--text-sm)", color: "var(--muted)", marginBottom: "var(--space-4)" }}>
        {t("taskBuilder.confirmHint")}
      </p>

      <div className="detail-header" style={{ marginBottom: "var(--space-5)" }}>
        <div className="detail-header-chips">
          <span className="detail-chip">
            {importMode === "execution" ? <Play size={13} className="detail-chip-icon" /> : <GitBranch size={13} className="detail-chip-icon" />}
            <span>{importMode === "execution" ? t("taskBuilder.byExecutionId") : t("taskBuilder.byTraceId")}</span>
          </span>
          <span className="detail-chip">
            <Hash size={13} className="detail-chip-icon" />
            <span style={{ fontFamily: "monospace", fontSize: "var(--text-xs)" }}>{id}</span>
          </span>
        </div>
      </div>

      <div style={{ background: "var(--surface-raised)", border: "1px solid var(--line)", borderRadius: "var(--radius-lg)", padding: "var(--space-4)" }}>
        <p style={{ fontSize: "var(--text-sm)", fontWeight: 600, color: "var(--ink-strong)", marginBottom: "var(--space-2)" }}>
          {t("taskBuilder.willExtract")}
        </p>
        <ul style={{ margin: 0, paddingLeft: "var(--space-5)", fontSize: "var(--text-sm)", color: "var(--ink)", lineHeight: 1.8 }}>
          <li>{t("taskBuilder.extractInput")}</li>
          <li>{t("taskBuilder.extractName")}</li>
          <li>{t("taskBuilder.extractDifficulty")}</li>
        </ul>
      </div>

      {error && <p className="form-error" style={{ marginTop: "var(--space-4)" }}>{error}</p>}
    </div>
  );
}

/* ========== Step 3: Result ========== */

function Step3Result({ task }: { task: Task }) {
  const t = useTranslations("tasks");
  const tc = useTranslations("common");

  return (
    <div>
      <div style={{ display: "flex", alignItems: "center", gap: "var(--space-3)", marginBottom: "var(--space-5)" }}>
        <CheckCircle size={20} style={{ color: "var(--success)", flexShrink: 0 }} />
        <span style={{ fontWeight: 600, color: "var(--success)", fontSize: "var(--text-sm)" }}>
          {t("taskBuilder.buildSuccess")}
        </span>
      </div>

      <header className="detail-header">
        <div className="detail-header-main">
          <h1 className="detail-header-title">{task.name}</h1>
          <Badge variant={task.difficulty as "easy" | "medium" | "hard"}>{task.difficulty}</Badge>
        </div>
        <div className="detail-header-chips">
          <span className="detail-chip">
            <Hash size={13} className="detail-chip-icon" />
            <span style={{ fontFamily: "monospace", fontSize: "var(--text-xs)" }}>{task.id.slice(0, 8)}…</span>
          </span>
          {task.test_cases && task.test_cases.length > 0 && (
            <span className="detail-chip">
              <FileText size={13} className="detail-chip-icon" />
              <span>{task.test_cases.length} test cases</span>
            </span>
          )}
        </div>
      </header>

      {task.description && (
        <p style={{ fontSize: "var(--text-sm)", color: "var(--muted)", marginBottom: "var(--space-4)" }}>
          {task.description}
        </p>
      )}

      {task.tags && task.tags.length > 0 && (
        <div className="tags" style={{ marginBottom: "var(--space-5)" }}>
          {task.tags.map((tag) => (
            <span key={tag} className="tag">{tag}</span>
          ))}
        </div>
      )}

      {task.test_cases && task.test_cases.length > 0 && (
        <div>
          <h3 style={{ fontSize: "var(--text-sm)", fontWeight: 600, color: "var(--ink-strong)", marginBottom: "var(--space-3)" }}>
            {t("taskBuilder.extractedTestCases")} ({task.test_cases.length})
          </h3>
          <div className="gateway-table-wrapper">
            <table className="gateway-table">
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
                        fontSize: "var(--text-xs)",
                        fontFamily: "monospace",
                        background: "var(--bg)",
                        padding: "var(--space-2)",
                        borderRadius: "var(--radius-sm)",
                      }}>
                        {typeof tc.input === "string" ? tc.input : JSON.stringify(tc.input, null, 2)}
                      </div>
                    </td>
                    <td>
                      <div style={{
                        maxWidth: 300,
                        maxHeight: 100,
                        overflow: "hidden",
                        textOverflow: "ellipsis",
                        whiteSpace: "pre-wrap",
                        fontSize: "var(--text-xs)",
                        fontFamily: "monospace",
                        background: "var(--bg)",
                        padding: "var(--space-2)",
                        borderRadius: "var(--radius-sm)",
                      }}>
                        {typeof tc.expected === "string" ? tc.expected : JSON.stringify(tc.expected, null, 2)}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  );
}
