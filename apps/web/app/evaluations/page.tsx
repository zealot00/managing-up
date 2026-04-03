import { Suspense } from "react";
import { getTaskExecutions, getTasks, getMetrics } from "../lib/api";
import EvaluationManager from "../components/EvaluationManager";

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

async function EvaluationsContent() {
  const [executions, tasks, metrics] = await Promise.all([
    getTaskExecutions().catch(() => null),
    getTasks().catch(() => null),
    getMetrics().catch(() => null),
  ]);

  const execItems = executions?.items ?? [];
  const taskItems = tasks?.items ?? [];
  const metricItems = metrics?.items ?? [];

  return (
    <main className="shell">
      <header className="hero-page hero-compact">
        <p className="eyebrow">Evaluation Engine</p>
        <h1>Task Executions</h1>
        <p className="lede">
          View agent performance across evaluation tasks. Each execution runs a task and produces metrics scores.
        </p>
      </header>

      <EvaluationManager
        executions={execItems}
        tasks={taskItems}
        metrics={metricItems}
      />
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
