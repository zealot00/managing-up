import { Suspense } from "react";
import { getTranslations } from "next-intl/server";
import { getSEHDashboardSummary, getSEHDatasets, getSEHRuns, getSEHPolicies } from "../lib/seh-api";
import SEHManager from "../components/SEHManager";
import { PageHeader } from "../components/layout/PageHeader";

async function SEHDashboardContent() {
  const t = await getTranslations("seh");
  const tc = await getTranslations("common");

  const results = await Promise.allSettled([
    getSEHDashboardSummary(),
    getSEHDatasets(50, 0),
    getSEHRuns(50, 0),
    getSEHPolicies(),
  ]);

  const [summaryResult, datasetsResult, runsResult, policiesResult] = results;
  const hasError = results.some(r => r.status === "rejected");

  const summary = summaryResult.status === "fulfilled"
    ? summaryResult.value
    : { total_datasets: 0, total_runs: 0, total_policies: 0, total_cases: 0, recent_runs: [], avg_score: 0, avg_success_rate: 0 };

  const datasets = datasetsResult.status === "fulfilled"
    ? datasetsResult.value.datasets
    : [];

  const runs = runsResult.status === "fulfilled"
    ? runsResult.value.runs
    : [];

  const policies = policiesResult.status === "fulfilled"
    ? policiesResult.value.map(p => ({ ...p, source_policies: p.source_policies ?? undefined }))
    : [];

  return (
    <main className="shell">
      <PageHeader
        eyebrow={t("eyebrow")}
        title={t("title")}
        description={t("lede")}
      />

      {hasError && (
        <div className="alert-bar alert-bar-warning">
          <span className="alert-bar-icon">⚠</span>
          <span className="alert-bar-text">{t("apiUnavailable")}</span>
          <span className="alert-bar-desc">{t("apiUnavailableDesc")}</span>
        </div>
      )}

      <SEHManager
        summary={summary}
        datasets={datasets}
        runs={runs}
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
