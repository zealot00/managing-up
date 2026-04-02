import { Suspense } from "react";
import { getTaskExecutions, getTasks, getMetrics } from "../lib/api";
import type { TaskExecution, Task, Metric } from "../lib/api";

function SkeletonEvaluations() {
  return (
    <main className="shell">
      <header className="hero-page hero-compact">
        <p className="eyebrow">Evaluation Engine</p>
        <h1>Task Executions</h1>
        <p className="lede">
          View agent performance across evaluation tasks. Each execution runs a task and produces metrics scores.
        </p>
      </header>
      <div className="skeleton-grid">
        {[1, 2, 3, 4].map((i) => (
          <div key={i} className="skeleton-card" />
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
    <article className="eval-card">
      <div className="eval-card-header">
        <div>
          <h3 className="eval-card-title">{taskName}</h3>
          <p className="eval-card-meta">Agent: {exec.agent_id}</p>
        </div>
        <span className={`badge badge-${exec.status === "completed" ? "succeeded" : exec.status === "failed" ? "failed" : "running"}`}>
          {exec.status}
        </span>
      </div>
      {exec.duration_ms && (
        <p className="eval-card-body" style={{ marginTop: "var(--space-3)" }}>
          Duration: {exec.duration_ms}ms
        </p>
      )}
      <div className="eval-card-footer">
        <span>Created {new Date(exec.created_at).toLocaleString()}</span>
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
      <header className="hero-page hero-compact">
        <p className="eyebrow">Evaluation Engine</p>
        <h1>Task Executions</h1>
        <p className="lede">
          View agent performance across evaluation tasks. Each execution runs a task and produces metrics scores.
        </p>
      </header>

      <section aria-label="Available metrics" style={{ marginTop: "var(--space-6)" }}>
        <h2 className="section-kicker" style={{ marginBottom: "var(--space-3)" }}>
          Available Metrics ({(metrics?.items ?? []).length})
        </h2>
        <div className="tags">
          {(metrics?.items ?? []).map((metric) => (
            <span key={metric.id} className="tag">
              {metric.name} ({metric.type})
            </span>
          ))}
          {(!metrics?.items || metrics.items.length === 0) && (
            <span style={{ color: "var(--muted)", fontSize: "var(--text-sm)" }}>
              No metrics defined. Create metrics via POST /api/v1/metrics.
            </span>
          )}
        </div>
      </section>

      <section aria-label="Task executions" style={{ marginTop: "var(--space-8)" }}>
        <h2 className="section-kicker" style={{ marginBottom: "var(--space-3)" }}>
          Task Executions ({(executions?.items ?? []).length})
        </h2>
        {(executions?.items ?? []).length > 0 ? (
          <div className="eval-grid">
            {executions?.items.map((exec) => (
              <ExecutionCard key={exec.id} exec={exec} tasks={tasks?.items ?? []} />
            ))}
          </div>
        ) : (
          <div className="empty-state">
            <div className="empty-state-icon">◎</div>
            <h3 className="empty-state-title">No executions yet</h3>
            <p className="empty-state-description">
              Task executions will appear here after running evaluations via POST /api/v1/task-executions.
            </p>
          </div>
        )}
      </section>

      <section aria-label="Task overview" style={{ marginTop: "var(--space-8)" }}>
        <h2 className="section-kicker" style={{ marginBottom: "var(--space-3)" }}>
          Task Overview ({(tasks?.items ?? []).length} tasks)
        </h2>
        {(tasks?.items ?? []).length > 0 ? (
          <div className="panel">
            <div className="list">
              {tasks?.items.map((task) => (
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
          </div>
        ) : (
          <p style={{ color: "var(--muted)", fontSize: "var(--text-sm)" }}>
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
