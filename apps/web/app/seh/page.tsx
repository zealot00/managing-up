import { Suspense } from "react";
import { getSEHDashboardSummary, getSEHDatasets, getSEHRuns, getSEHPolicies } from "../lib/seh-api";

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

async function SEHDashboardContent() {
  const summary = await getSEHDashboardSummary();
  const datasets = await getSEHDatasets(5, 0);
  const runs = await getSEHRuns(5, 0);
  const policies = await getSEHPolicies();

  const metrics = [
    { label: "Datasets", value: summary.total_datasets, icon: "□" },
    { label: "Evaluation Runs", value: summary.total_runs, icon: "▶" },
    { label: "Governance Policies", value: summary.total_policies, icon: "◎" },
    { label: "Avg Score", value: summary.avg_score > 0 ? summary.avg_score.toFixed(2) : "—", icon: "◉" },
  ];

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
        {metrics.map((metric) => (
          <article className="metric-card" key={metric.label}>
            <div className="metric-card-icon">{metric.icon}</div>
            <div className="metric-card-value">{metric.value}</div>
            <div className="metric-card-label">{metric.label}</div>
          </article>
        ))}
      </div>

      <div className="grid" style={{ marginTop: "var(--space-6)" }}>
        <section className="panel">
          <div className="panel-header">
            <p className="section-kicker">Datasets</p>
            <h2 className="panel-title">Recent Datasets</h2>
          </div>
          <div className="list">
            {datasets.datasets.length === 0 ? (
              <p className="empty-note">No datasets found</p>
            ) : (
              datasets.datasets.map((ds) => (
                <article className="list-card" key={ds.dataset_id}>
                  <div className="list-card-main">
                    <h3 className="list-card-title">{ds.name}</h3>
                    <p className="list-card-meta">
                      {ds.version} · {ds.owner} · {ds.case_count} cases
                    </p>
                  </div>
                  <span className="badge badge-muted">{ds.dataset_id}</span>
                </article>
              ))
            )}
          </div>
        </section>

        <section className="panel">
          <div className="panel-header">
            <p className="section-kicker">Performance</p>
            <h2 className="panel-title">Recent Runs</h2>
          </div>
          <div className="list">
            {runs.runs.length === 0 ? (
              <p className="empty-note">No runs found</p>
            ) : (
              runs.runs.map((run) => (
                <article className="list-card" key={run.run_id}>
                  <div className="list-card-main">
                    <h3 className="list-card-title">{run.skill}</h3>
                    <p className="list-card-meta">
                      Score: {run.metrics.score.toFixed(2)} · Success: {(run.metrics.success_rate * 100).toFixed(0)}%
                    </p>
                  </div>
                  <span className={`badge badge-${run.metrics.score >= 0.75 ? "succeeded" : "failed"}`}>
                    {run.run_id}
                  </span>
                </article>
              ))
            )}
          </div>
        </section>
      </div>

      <section className="panel" style={{ marginTop: "var(--space-6)" }}>
        <div className="panel-header">
          <p className="section-kicker">Governance</p>
          <h2 className="panel-title">Active Policies</h2>
        </div>
        <div className="list">
          {policies.length === 0 ? (
            <p className="empty-note">No policies found</p>
          ) : (
            policies.map((policy) => (
              <article className="list-card" key={policy.policy_id}>
                <div className="list-card-main">
                  <h3 className="list-card-title">{policy.name}</h3>
                  <p className="list-card-meta">
                    {policy.require_provenance ? "Provenance required · " : ""}
                    Min diversity: {policy.min_source_diversity} · Min golden: {policy.min_golden_weight}
                  </p>
                </div>
                <span className="badge badge-muted">{policy.policy_id}</span>
              </article>
            ))
          )}
        </div>
      </section>
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
