import { Suspense } from "react";
import { getTaskExecutions, getTasks, getMetrics } from "../lib/api";
import type { TaskExecution, Task, Metric } from "../lib/api";

function SkeletonEvaluations() {
  return (
    <main className="shell">
      <section className="toprail">
        <div className="loading-pulse" style={{ width: 180, height: 44, borderRadius: 999 }} />
      </section>
      <div className="loading-pulse loading-pulse-medium" style={{ marginBottom: 8 }} />
      <div className="skeleton-grid">
        {[1, 2, 3, 4].map((i) => (
          <div className="skeleton-card" key={i} />
        ))}
      </div>
    </main>
  );
}

function getScoreClass(score: number): string {
  if (score >= 0.8) return "score-fill score-fill-high";
  if (score >= 0.5) return "score-fill score-fill-medium";
  return "score-fill score-fill-low";
}

function ExecutionCard({ exec, tasks }: { exec: TaskExecution; tasks: Task[] }) {
  const task = tasks.find((t) => t.id === exec.task_id);
  const taskName = task?.name || exec.task_id;

  return (
    <article className="card">
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", gap: 12 }}>
        <div>
          <h2 style={{ margin: "0 0 4px", fontSize: "1.1rem" }}>{taskName}</h2>
          <p style={{ margin: 0, color: "var(--muted)", fontSize: "0.82rem" }}>
            Agent: {exec.agent_id}
          </p>
        </div>
        <span className={`badge badge-${exec.status === "completed" ? "succeeded" : exec.status === "failed" ? "failed" : "running"}`}>
          {exec.status}
        </span>
      </div>
      {exec.duration_ms && (
        <p style={{ margin: "12px 0 0", fontSize: "0.82rem", color: "var(--muted)" }}>
          Duration: {exec.duration_ms}ms
        </p>
      )}
      <div style={{ marginTop: 12, paddingTop: 12, borderTop: "1px solid var(--line)" }}>
        <p style={{ margin: "0 0 4px", fontSize: "0.78rem", color: "var(--muted)", textTransform: "uppercase", letterSpacing: "0.04em" }}>
          Created
        </p>
        <p style={{ margin: 0, fontSize: "0.85rem" }}>
          {new Date(exec.created_at).toLocaleString()}
        </p>
      </div>
    </article>
  );
}

async function EvaluationsContent() {
  let executions: { items: TaskExecution[] } | null = null;
  let tasks: { items: Task[] } | null = null;
  let metrics: { items: Metric[] } | null = null;

  try {
    [executions, tasks, metrics] = await Promise.all([
      getTaskExecutions(),
      getTasks(),
      getMetrics(),
    ]);
  } catch {
    executions = null;
    tasks = null;
    metrics = null;
  }

  const taskMap = new Map((tasks?.items ?? []).map((t) => [t.id, t]));

  return (
    <main className="shell">
      <section className="toprail">
        <a className="toprail-link" href="/">
          Dashboard
        </a>
        <a className="toprail-link" href="/tasks">
          Tasks
        </a>
        <a className="toprail-link" href="/evaluations">
          Evaluations
        </a>
      </section>

      <section className="hero-page hero-compact">
        <p className="eyebrow">Evaluation Engine</p>
        <h1>Task Executions</h1>
        <p className="lede">
          View agent performance across evaluation tasks. Each execution runs a task
          and produces metrics scores.
        </p>
      </section>

      <section aria-label="Available metrics" style={{ marginTop: 24 }}>
        <h2 style={{ fontSize: "1rem", color: "var(--muted)", textTransform: "uppercase", letterSpacing: "0.08em", marginBottom: 12 }}>
          Available Metrics ({(metrics?.items ?? []).length})
        </h2>
        <div style={{ display: "flex", gap: 8, flexWrap: "wrap" }}>
          {(metrics?.items ?? []).map((metric) => (
            <span key={metric.id} className="badge badge-muted">
              {metric.name} ({metric.type})
            </span>
          ))}
          {(!metrics?.items || metrics.items.length === 0) && (
            <span style={{ color: "var(--muted)", fontSize: "0.85rem" }}>
              No metrics defined. Create metrics via POST /api/v1/metrics.
            </span>
          )}
        </div>
      </section>

      <section aria-label="Task executions" style={{ marginTop: 32 }}>
        <h2 style={{ fontSize: "1rem", color: "var(--muted)", textTransform: "uppercase", letterSpacing: "0.08em", marginBottom: 12 }}>
          Task Executions ({(executions?.items ?? []).length})
        </h2>
        {(executions?.items ?? []).length > 0 ? (
          <div className="eval-grid">
            {executions?.items.map((exec) => (
              <ExecutionCard key={exec.id} exec={exec} tasks={tasks?.items ?? []} />
            ))}
          </div>
        ) : (
          <article className="panel">
            <div className="panel-header">
              <h2>No executions yet</h2>
            </div>
            <p style={{ color: "var(--muted)", marginTop: 12 }}>
              Task executions will appear here after running evaluations via POST /api/v1/task-executions.
            </p>
          </article>
        )}
      </section>

      <section aria-label="Task overview" style={{ marginTop: 32 }}>
        <h2 style={{ fontSize: "1rem", color: "var(--muted)", textTransform: "uppercase", letterSpacing: "0.08em", marginBottom: 12 }}>
          Task Overview ({(tasks?.items ?? []).length} tasks)
        </h2>
        {(tasks?.items ?? []).length > 0 ? (
          <article className="panel">
            <div className="list">
              {tasks?.items.map((task) => (
                <article className="list-card" key={task.id}>
                  <div>
                    <h3 style={{ margin: "0 0 4px", fontSize: "1rem" }}>{task.name}</h3>
                    <p style={{ margin: 0, fontSize: "0.82rem" }}>
                      {task.test_cases.length} test cases · {task.difficulty} difficulty
                    </p>
                  </div>
                  <span className={`badge badge-${task.difficulty === "easy" ? "succeeded" : task.difficulty === "medium" ? "running" : "failed"}`}>
                    {task.difficulty}
                  </span>
                </article>
              ))}
            </div>
          </article>
        ) : (
          <p style={{ color: "var(--muted)", fontSize: "0.85rem" }}>
            No tasks defined. Create tasks via POST /api/v1/tasks.
          </p>
        )}
      </section>
    </main>
  );
}

export default function EvaluationsPage() {
  return (
    <Suspense fallback={<SkeletonEvaluations />}>
      <EvaluationsContent />
    </Suspense>
  );
}
