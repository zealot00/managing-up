"use client";

import { useState, useMemo } from "react";
import { useRouter } from "next/navigation";
import { TaskExecution, Task, Metric, runTaskEvaluation, createMetric } from "../lib/api";
import { useTranslations } from "next-intl";
import RunEvaluationForm from "./RunEvaluationForm";
import CreateMetricForm from "./CreateMetricForm";
import { PageHeader } from "./layout/PageHeader";
import { EmptyState } from "./layout/EmptyState";
import { Badge } from "./ui/Badge";
import { DataToolbar } from "./ui/DataToolbar";

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
          onSearchChange={setSearchQuery}
          filters={
            <>
              <select
                value={statusFilter}
                onChange={(e) => setStatusFilter(e.target.value)}
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
            <h2 className="panel-title">{t("taskExecutions", { count: filteredExecutions.length })}</h2>
          </div>
          {filteredExecutions.length > 0 ? (
            <div className="eval-grid">
              {filteredExecutions.map((exec) => (
                <ExecutionCard key={exec.id} exec={exec} tasks={tasks} />
              ))}
            </div>
          ) : (
            <EmptyState
              icon="◎"
              title={t("noExecutions")}
              description={t("noExecutionsDesc")}
            />
          )}
        </div>
      </section>

      <section aria-label="Task overview" style={{ marginTop: "var(--space-6)" }}>
        <DataToolbar
          searchQuery={searchQuery}
          onSearchChange={setSearchQuery}
          filters={
            <select
              value={difficultyFilter}
              onChange={(e) => setDifficultyFilter(e.target.value)}
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
            <h2 className="panel-title">{t("taskOverview", { count: filteredTasks.length })}</h2>
          </div>
          {filteredTasks.length > 0 ? (
            <div className="list">
              {filteredTasks.map((task) => (
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
        <Badge variant={exec.status as "running" | "pending" | "completed" | "failed" | "muted"}>
          {exec.status}
        </Badge>
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
