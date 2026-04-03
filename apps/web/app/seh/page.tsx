import { Suspense } from "react";
import { getTranslations } from "next-intl/server";
import { getSEHDashboardSummary, getSEHDatasets, getSEHRuns, getSEHPolicies } from "../lib/seh-api";
import SEHManager from "../components/SEHManager";

async function SEHDashboardContent() {
  const t = await getTranslations("seh");
  const summary = await getSEHDashboardSummary();
  const datasets = await getSEHDatasets(50, 0);
  const runs = await getSEHRuns(50, 0);
  const policies = await getSEHPolicies();

  return (
    <main className="shell">
      <header className="hero-page hero-compact">
        <p className="eyebrow">{t("eyebrow")}</p>
        <h1>{t("title")}</h1>
        <p className="lede">
          {t("lede")}
        </p>
      </header>

      <SEHManager
        summary={summary}
        datasets={datasets.datasets}
        runs={runs.runs}
        policies={policies}
      />
    </main>
  );
}

function SkeletonSEHDashboard() {
  return (
    <main className="shell">
      <header className="hero-page hero-compact">
        <p className="eyebrow">Skill Evaluation Hub</p>
        <h1>SEH Dashboard</h1>
        <p className="lede">
          Monitoring and management for datasets, evaluation runs, and governance policies.
        </p>
      </header>

      <div className="stats">
        {[...Array(4)].map((_, i) => (
          <div key={i} className="metric-card">
            <div className="loading-pulse loading-pulse-short" style={{ width: 80, marginBottom: 8 }} />
            <div className="loading-pulse" style={{ width: 60, height: 32 }} />
          </div>
        ))}
      </div>

      <div className="grid">
        <div className="panel">
          <div className="loading-pulse loading-pulse-medium" style={{ marginBottom: 16 }} />
          <div className="skeleton-grid">
            {[1, 2, 3].map((i) => <div key={i} className="skeleton-card" />)}
          </div>
        </div>
        <div className="panel">
          <div className="loading-pulse loading-pulse-medium" style={{ marginBottom: 16 }} />
          <div className="skeleton-grid">
            {[1, 2, 3].map((i) => <div key={i} className="skeleton-card" />)}
          </div>
        </div>
      </div>
    </main>
  );
}

export default function SEHDashboardPage() {
  return (
    <Suspense fallback={<SkeletonSEHDashboard />}>
      <SEHDashboardContent />
    </Suspense>
  );
}