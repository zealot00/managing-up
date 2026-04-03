import { Suspense } from "react";
import { getTranslations } from "next-intl/server";
import { getExperiments, getTasks } from "../lib/api";
import ExperimentManager from "../components/ExperimentManager";

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
      <header className="hero-page hero-compact">
        <p className="eyebrow">{t("eyebrow")}</p>
        <h1>{t("title")}</h1>
        <p className="lede">
          {t("lede")}
        </p>
      </header>

      <ExperimentManager experiments={experiments} tasks={tasks} />
    </main>
  );
}

function SkeletonExperiments() {
  return (
    <main className="shell">
      <header className="hero-page hero-compact">
        <p className="eyebrow">Experiment DB</p>
        <h1>Experiments</h1>
        <p className="lede">
          Compare agent and skill performance across multiple task executions.
        </p>
      </header>
      <div className="skeleton-grid">
        {[1, 2, 3].map((i) => (
          <div key={i} className="skeleton-card" />
        ))}
      </div>
    </main>
  );
}

export default function ExperimentsPage() {
  return (
    <Suspense fallback={<SkeletonExperiments />}>
      <ExperimentsContent />
    </Suspense>
  );
}