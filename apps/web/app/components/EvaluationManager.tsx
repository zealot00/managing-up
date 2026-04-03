"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { TaskExecution, Task, Metric, runTaskEvaluation, createMetric } from "../lib/api";
import { useTranslations } from "next-intl";
import RunEvaluationForm from "./RunEvaluationForm";
import CreateMetricForm from "./CreateMetricForm";

type Props = {
  executions: TaskExecution[];
  tasks: Task[];
  metrics: Metric[];
};

export default function EvaluationManager({ executions, tasks, metrics }: Props) {
  const t = useTranslations("evaluations");
  const [showRunEval, setShowRunEval] = useState(false);
  const [showCreateMetric, setShowCreateMetric] = useState(false);

  return (
    <>
      <div className="page-header" style={{ marginBottom: "var(--space-6)", marginTop: "var(--space-4)", paddingBottom: 0, borderBottom: "none" }}>
        <div className="page-header-content">
          <p className="section-kicker" style={{ margin: 0 }}>
            {t("taskExecutions", { count: executions.length })}
          </p>
        </div>
        <div className="page-header-actions">
          <button className="btn btn-secondary" onClick={() => { setShowCreateMetric(!showCreateMetric); setShowRunEval(false); }}>
            {showCreateMetric ? "Cancel" : t("newMetric")}
          </button>
          <button className="btn btn-primary" onClick={() => { setShowRunEval(!showRunEval); setShowCreateMetric(false); }}>
            {showRunEval ? "Cancel" : t("runEvaluation")}
          </button>
        </div>
      </div>

      {showCreateMetric && <CreateMetricForm onCreated={() => setShowCreateMetric(false)} />}
      {showRunEval && <RunEvaluationForm tasks={tasks} onCreated={() => setShowRunEval(false)} />}

      <section aria-label="Available metrics" style={{ marginTop: "var(--space-6)" }}>
        <div className="panel">
          <div className="panel-header">
            <p className="section-kicker">{t("metrics")}</p>
            <h2 className="panel-title">{t("availableMetrics", { count: metrics.length })}</h2>
          </div>
          {metrics.length > 0 ? (
            <div className="tags">
              {metrics.map((metric) => (
                <span key={metric.id} className="tag">
                  {metric.name} <span style={{ opacity: 0.6 }}>({metric.type})</span>
                </span>
              ))}
            </div>
          ) : (
            <p className="empty-note">{t("noMetrics")}</p>
          )}
        </div>
      </section>

      <section aria-label="Task executions" style={{ marginTop: "var(--space-6)" }}>
        <div className="panel">
          <div className="panel-header">
            <p className="section-kicker">{t("eyebrow")}</p>
            <h2 className="panel-title">{t("taskExecutions", { count: executions.length })}</h2>
          </div>
          {executions.length > 0 ? (
            <div className="eval-grid">
              {executions.map((exec) => (
                <ExecutionCard key={exec.id} exec={exec} tasks={tasks} />
              ))}
            </div>
          ) : (
            <div className="empty-state">
              <div className="empty-state-icon">◎</div>
              <h3 className="empty-state-title">{t("noExecutions")}</h3>
              <p className="empty-state-description">{t("noExecutionsDesc")}</p>
            </div>
          )}
        </div>
      </section>

      <section aria-label="Task overview" style={{ marginTop: "var(--space-6)" }}>
        <div className="panel">
          <div className="panel-header">
            <p className="section-kicker">Tasks</p>
            <h2 className="panel-title">{t("taskOverview", { count: tasks.length })}</h2>
          </div>
          {tasks.length > 0 ? (
            <div className="list">
              {tasks.map((task) => (
                <article className="list-card" key={task.id}>
                  <div className="list-card-main">
                    <h3 className="list-card-title">{task.name}</h3>
                    <p className="list-card-meta">
                      {task.test_cases.length} test cases · {task.difficulty} difficulty
                    </p>
                  </div>
                  <span className={`badge badge-${task.difficulty === "easy" ? "succeeded" : task.difficulty === "medium" ? "running" : "failed"}`}>
                    {task.difficulty}
                  </span>
                </article>
              ))}
            </div>
          ) : (
            <p className="empty-note">{t("noTasksDesc")}</p>
          )}
        </div>
      </section>
    </>
  );
}

function ExecutionCard({ exec, tasks }: { exec: TaskExecution; tasks: Task[] }) {
  const t = useTranslations("evaluations");
  const task = tasks.find((t) => t.id === exec.task_id);
  const taskName = task?.name || exec.task_id;

  return (
    <article className="eval-card">
      <div className="eval-card-header">
        <div>
          <h3 className="eval-card-title">{taskName}</h3>
          <p className="eval-card-meta">{t("agent")}: {exec.agent_id}</p>
        </div>
        <span className={`badge badge-${exec.status === "completed" ? "succeeded" : exec.status === "failed" ? "failed" : "running"}`}>
          {exec.status}
        </span>
      </div>
      {exec.duration_ms && (
        <p className="eval-card-body" style={{ marginTop: "var(--space-3)" }}>
          {t("duration", { ms: exec.duration_ms })}
        </p>
      )}
      <div className="eval-card-footer">
        <span>{t("createdAt")} {new Date(exec.created_at).toLocaleString()}</span>
      </div>
    </article>
  );
}
