import { Suspense } from "react";
import { getTranslations } from "next-intl/server";
import { getTaskExecutions, getTasks, getMetrics } from "../lib/api";
import EvaluationManager from "../components/EvaluationManager";

async function EvaluationsContent() {
  const t = await getTranslations("evaluations");
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
        <p className="eyebrow">{t("eyebrow")}</p>
        <h1>{t("title")}</h1>
        <p className="lede">
          {t("lede")}
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

export default function EvaluationsPage() {
  return (
    <Suspense fallback={<SkeletonEvaluations />}>
      <EvaluationsContent />
    </Suspense>
  );
}