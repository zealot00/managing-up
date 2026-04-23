"use client";

import { useState, useMemo } from "react";
import { TaskExecution, Task, Metric } from "../lib/api";
import { useTranslations } from "next-intl";
import RunEvaluationForm from "./RunEvaluationForm";
import CreateMetricForm from "./CreateMetricForm";
import { PageHeader } from "./layout/PageHeader";
import { EmptyState } from "./layout/EmptyState";
import { Badge } from "./ui/Badge";
import { formatRelativeTime, formatDurationMs } from "../lib/format";

const PAGE_SIZE = 20;

type Props = {
  executions: TaskExecution[];
  tasks: Task[];
  metrics: Metric[];
};

export default function EvaluationManager({ executions, tasks, metrics }: Props) {
  const t = useTranslations("evaluations");
  const [showRunEval, setShowRunEval] = useState(false);
  const [showCreateMetric, setShowCreateMetric] = useState(false);

  const [searchQuery, setSearchQuery] = useState("");
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const [difficultyFilter, setDifficultyFilter] = useState<string>("all");
  const [execDisplayCount, setExecDisplayCount] = useState(PAGE_SIZE);
  const [taskDisplayCount, setTaskDisplayCount] = useState(PAGE_SIZE);

  const filteredExecutions = useMemo(() => {
    return executions.filter((exec) => {
      if (searchQuery) {
        const task = tasks.find((t) => t.id === exec.task_id);
        const taskName = task?.name || "";
        if (!taskName.toLowerCase().includes(searchQuery.toLowerCase()) &&
            !exec.agent_id.toLowerCase().includes(searchQuery.toLowerCase())) {
          return false;
        }
      }
      if (statusFilter !== "all" && exec.status !== statusFilter) {
        return false;
      }
      return true;
    });
  }, [executions, tasks, searchQuery, statusFilter]);

  const filteredTasks = useMemo(() => {
    return tasks.filter((task) => {
      if (searchQuery) {
        if (!task.name.toLowerCase().includes(searchQuery.toLowerCase()) &&
            !task.description?.toLowerCase().includes(searchQuery.toLowerCase())) {
          return false;
        }
      }
      if (difficultyFilter !== "all" && task.difficulty !== difficultyFilter) {
        return false;
      }
      return true;
    });
  }, [tasks, searchQuery, difficultyFilter]);

  const displayedExecutions = filteredExecutions.slice(0, execDisplayCount);
  const displayedTasks = filteredTasks.slice(0, taskDisplayCount);
  const hasMoreExec = execDisplayCount < filteredExecutions.length;
  const hasMoreTasks = taskDisplayCount < filteredTasks.length;

  const stats = [
    { label: t("taskExecutions", { count: executions.length }), value: executions.length, icon: "▶", color: "var(--primary)" },
    { label: t("tasks", { count: tasks.length }), value: tasks.length, icon: "☐", color: "var(--ink)" },
    { label: t("metrics", { count: metrics.length }), value: metrics.length, icon: "◆", color: "var(--success)" },
    { label: "Running", value: executions.filter(e => e.status === "running").length, icon: "●", color: "var(--warning)" },
  ];

  return (
    <>
      <PageHeader
        title={t("title")}
        description={t("lede")}
        actions={
          <>
            <button className="btn btn-secondary" onClick={() => { setShowCreateMetric(!showCreateMetric); setShowRunEval(false); }}>
              {showCreateMetric ? "Cancel" : t("newMetric")}
            </button>
            <button className="btn btn-primary" onClick={() => { setShowRunEval(!showRunEval); setShowCreateMetric(false); }}>
              {showRunEval ? "Cancel" : t("runEvaluation")}
            </button>
          </>
        }
      />

      {showCreateMetric && <CreateMetricForm onCreated={() => setShowCreateMetric(false)} />}
      {showRunEval && <RunEvaluationForm tasks={tasks} onCreated={() => setShowRunEval(false)} />}

      <div className="dashboard-stats">
        {stats.map((stat) => (
          <article className="dashboard-stat-card" key={stat.label}>
            <div className="dashboard-stat-icon" style={{ color: stat.color }}>{stat.icon}</div>
            <div className="dashboard-stat-value">{stat.value}</div>
            <div className="dashboard-stat-label">{stat.label}</div>
          </article>
        ))}
      </div>

      {metrics.length > 0 && (
        <div className="dashboard-section" style={{ marginBottom: "var(--space-6)" }}>
          <div className="dashboard-section-header">
            <h2 className="dashboard-section-title">{t("availableMetrics", { count: metrics.length })}</h2>
          </div>
          <div className="tags">
            {metrics.map((metric) => (
              <span key={metric.id} className="tag">
                {metric.name} <span style={{ opacity: 0.6 }}>({metric.type})</span>
              </span>
            ))}
          </div>
        </div>
      )}

      <div className="content-grid">
        <div className="dashboard-section">
          <div className="dashboard-section-header">
            <h2 className="dashboard-section-title">
              {t("eyebrow")}
            </h2>
            <div style={{ display: "flex", gap: "var(--space-2)" }}>
              <select
                value={statusFilter}
                onChange={(e) => { setStatusFilter(e.target.value); setExecDisplayCount(PAGE_SIZE); }}
                className="form-select"
                style={{ minWidth: 120, fontSize: "var(--text-sm)" }}
              >
                <option value="all">All Status</option>
                <option value="running">Running</option>
                <option value="pending">Pending</option>
                <option value="completed">Completed</option>
                <option value="failed">Failed</option>
              </select>
            </div>
          </div>
          <div style={{ marginBottom: "var(--space-4)" }}>
            <input
              type="text"
              placeholder="Search executions..."
              value={searchQuery}
              onChange={(e) => { setSearchQuery(e.target.value); setExecDisplayCount(PAGE_SIZE); }}
              className="form-input"
              style={{ width: "100%" }}
            />
          </div>
          {displayedExecutions.length > 0 ? (
            <>
              <div className="dashboard-list">
                {displayedExecutions.map((exec) => (
                  <ExecutionRow key={exec.id} exec={exec} tasks={tasks} />
                ))}
              </div>
              {hasMoreExec && (
                <button
                  onClick={() => setExecDisplayCount((c) => c + PAGE_SIZE)}
                  className="btn btn-ghost btn-sm"
                  style={{ width: "100%", marginTop: "var(--space-3)" }}
                >
                  Load more ({filteredExecutions.length - execDisplayCount} remaining)
                </button>
              )}
            </>
          ) : (
            <EmptyState title={executions.length > 0 ? "No matching executions" : t("noExecutions")} />
          )}
        </div>

        <div className="dashboard-section">
          <div className="dashboard-section-header">
            <h2 className="dashboard-section-title">
              {t("taskOverview", { count: tasks.length })}
            </h2>
            <div style={{ display: "flex", gap: "var(--space-2)" }}>
              <select
                value={difficultyFilter}
                onChange={(e) => { setDifficultyFilter(e.target.value); setTaskDisplayCount(PAGE_SIZE); }}
                className="form-select"
                style={{ minWidth: 120, fontSize: "var(--text-sm)" }}
              >
                <option value="all">All Difficulties</option>
                <option value="easy">Easy</option>
                <option value="medium">Medium</option>
                <option value="hard">Hard</option>
              </select>
            </div>
          </div>
          {displayedTasks.length > 0 ? (
            <>
              <div className="dashboard-list">
                {displayedTasks.map((task) => (
                  <article className="dashboard-list-item" key={task.id}>
                    <div className="dashboard-list-main">
                      <h3 className="dashboard-list-title">{task.name}</h3>
                      <p className="dashboard-list-meta">
                        {task.test_cases.length} test cases
                      </p>
                    </div>
                    <Badge variant={task.difficulty as "easy" | "medium" | "hard"}>
                      {task.difficulty}
                    </Badge>
                  </article>
                ))}
              </div>
              {hasMoreTasks && (
                <button
                  onClick={() => setTaskDisplayCount((c) => c + PAGE_SIZE)}
                  className="btn btn-ghost btn-sm"
                  style={{ width: "100%", marginTop: "var(--space-3)" }}
                >
                  Load more ({filteredTasks.length - taskDisplayCount} remaining)
                </button>
              )}
            </>
          ) : (
            <EmptyState title={tasks.length > 0 ? "No matching tasks" : t("noTasksDesc")} />
          )}
        </div>
      </div>
    </>
  );
}

function ExecutionRow({ exec, tasks }: { exec: TaskExecution; tasks: Task[] }) {
  const t = useTranslations("evaluations");
  const task = tasks.find((t) => t.id === exec.task_id);
  const taskName = task?.name || exec.task_id;

  return (
    <article className="list-card">
      <div className="list-card-main">
        <h3 className="list-card-title">{taskName}</h3>
        <p className="list-card-meta">
          {t("agent")}: {exec.agent_id}
          {exec.duration_ms && ` · ${formatDurationMs(exec.duration_ms)}`}
        </p>
      </div>
      <div className="list-card-actions">
        <span>{formatRelativeTime(exec.created_at)}</span>
        <Badge variant={exec.status as "running" | "pending" | "completed" | "failed" | "muted"}>
          {exec.status}
        </Badge>
      </div>
    </article>
  );
}
