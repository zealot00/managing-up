import { Suspense } from "react";
import { getTranslations } from "next-intl/server";
import { getTaskExecutions, getTasks, getMetrics } from "../lib/api";
import EvaluationManager from "../components/EvaluationManager";
import { PageSkeleton } from "../components/layout/Skeleton";
import { PageHeader } from "../components/layout/PageHeader";

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
      <EvaluationManager
        executions={execItems}
        tasks={taskItems}
        metrics={metricItems}
      />
    </main>
  );
}

function SkeletonEvaluations() {
  return <PageSkeleton headerActions={true} content="cards" contentCount={4} />;
}

export default function EvaluationsPage() {
  return (
    <Suspense fallback={<SkeletonEvaluations />}>
      <EvaluationsContent />
    </Suspense>
  );
}