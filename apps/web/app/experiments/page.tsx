import { Suspense } from "react";
import { getTranslations } from "next-intl/server";
import { getExperiments, getTasks } from "../lib/api";
import ExperimentManager from "../components/ExperimentManager";
import { PageSkeleton } from "../components/layout/Skeleton";
import { PageHeader } from "../components/layout/PageHeader";

async function ExperimentsContent() {
  const t = await getTranslations("experiments");
  const [experimentsResp, tasksResp] = await Promise.all([
    getExperiments().catch(() => null),
    getTasks().catch(() => null),
  ]);

  const experiments = experimentsResp?.items ?? [];
  const tasks = tasksResp?.items ?? [];

  return (
    <main className="shell">
      <PageHeader
        eyebrow={t("eyebrow")}
        title={t("title")}
        description={t("lede")}
      />

      <ExperimentManager experiments={experiments} tasks={tasks} />
    </main>
  );
}

function SkeletonExperiments() {
  return <PageSkeleton headerActions={true} content="cards" contentCount={3} />;
}

export default function ExperimentsPage() {
  return (
    <Suspense fallback={<SkeletonExperiments />}>
      <ExperimentsContent />
    </Suspense>
  );
}