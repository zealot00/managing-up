"use client";

import { useState, useMemo } from "react";
import { TaskExecution, Task, Metric } from "../lib/api";
import { useTranslations } from "next-intl";
import RunEvaluationForm from "./RunEvaluationForm";
import CreateMetricForm from "./CreateMetricForm";
import { PageHeader } from "./layout/PageHeader";
import { EmptyState } from "./layout/EmptyState";
import { Badge } from "./ui/Badge";
import { DataToolbar } from "./ui/DataToolbar";
import { LoadMore } from "./ui/LoadMore";
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

  return (
    <>
      <PageHeader
        eyebrow={t("taskExecutions", { count: executions.length })}
        title=""
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
        <DataToolbar
          searchQuery={searchQuery}
          onSearchChange={(v) => { setSearchQuery(v); setExecDisplayCount(PAGE_SIZE); }}
          filters={
            <>
              <select
                value={statusFilter}
                onChange={(e) => { setStatusFilter(e.target.value); setExecDisplayCount(PAGE_SIZE); }}
                className="form-select"
                style={{ minWidth: 140 }}
              >
                <option value="all">All Status</option>
                <option value="running">Running</option>
                <option value="pending">Pending</option>
                <option value="completed">Completed</option>
                <option value="failed">Failed</option>
              </select>
            </>
          }
        />
        <div className="panel">
          <div className="panel-header">
            <p className="section-kicker">{t("eyebrow")}</p>
            <h2 className="panel-title">
              {filteredExecutions.length === executions.length
                ? t("taskExecutions", { count: executions.length })
                : `${filteredExecutions.length} of ${executions.length} executions`}
            </h2>
          </div>
          {displayedExecutions.length > 0 ? (
            <>
              <div className="list">
                {displayedExecutions.map((exec) => (
                  <ExecutionRow key={exec.id} exec={exec} tasks={tasks} />
                ))}
              </div>
              <LoadMore
                hasMore={hasMoreExec}
                isLoading={false}
                onLoadMore={() => setExecDisplayCount((c) => c + PAGE_SIZE)}
                label="Load more executions"
              />
            </>
          ) : executions.length > 0 ? (
            <EmptyState title="No matching executions" />
          ) : (
            <EmptyState title={t("noExecutions")} description={t("noExecutionsDesc")} />
          )}
        </div>
      </section>

      <section aria-label="Task overview" style={{ marginTop: "var(--space-6)" }}>
        <DataToolbar
          searchQuery={searchQuery}
          onSearchChange={(v) => { setSearchQuery(v); setTaskDisplayCount(PAGE_SIZE); }}
          filters={
            <select
              value={difficultyFilter}
              onChange={(e) => { setDifficultyFilter(e.target.value); setTaskDisplayCount(PAGE_SIZE); }}
              className="form-select"
              style={{ minWidth: 140 }}
            >
              <option value="all">All Difficulties</option>
              <option value="easy">Easy</option>
              <option value="medium">Medium</option>
              <option value="hard">Hard</option>
            </select>
          }
        />
        <div className="panel">
          <div className="panel-header">
            <p className="section-kicker">Tasks</p>
            <h2 className="panel-title">
              {filteredTasks.length === tasks.length
                ? t("taskOverview", { count: tasks.length })
                : `${filteredTasks.length} of ${tasks.length} tasks`}
            </h2>
          </div>
          {displayedTasks.length > 0 ? (
            <>
              <div className="list">
                {displayedTasks.map((task) => (
                  <article className="list-card" key={task.id}>
                    <div className="list-card-main">
                      <h3 className="list-card-title">{task.name}</h3>
                      <p className="list-card-meta">
                        {task.test_cases.length} test cases · {task.difficulty} difficulty
                      </p>
                    </div>
                    <Badge variant={task.difficulty as "easy" | "medium" | "hard"}>
                      {task.difficulty}
                    </Badge>
                  </article>
                ))}
              </div>
              <LoadMore
                hasMore={hasMoreTasks}
                isLoading={false}
                onLoadMore={() => setTaskDisplayCount((c) => c + PAGE_SIZE)}
                label="Load more tasks"
              />
            </>
          ) : tasks.length > 0 ? (
            <EmptyState title="No matching tasks" />
          ) : (
            <p className="empty-note">{t("noTasksDesc")}</p>
          )}
        </div>
      </section>
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
