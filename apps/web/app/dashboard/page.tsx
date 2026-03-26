import { Suspense } from "react";
import { getDashboard } from "../lib/api";
import { SkeletonPanel } from "../components/SkeletonPanel";

function formatDuration(seconds: number): string {
  if (seconds < 60) {
    return `${seconds}s`;
  }
  const minutes = Math.floor(seconds / 60);
  const remainingSeconds = seconds % 60;
  return remainingSeconds > 0 ? `${minutes}m ${remainingSeconds}s` : `${minutes}m`;
}

function SkeletonDashboardPage() {
  return (
    <main className="shell">
      <section className="hero-page hero-compact">
        <p className="eyebrow">Observability</p>
        <h1>Skill hub at a glance.</h1>
        <p className="lede">
          Summary metrics and recent activity across your governed skill ecosystem.
        </p>
      </section>

      <div style={{ display: "grid", gridTemplateColumns: "repeat(3, 1fr)", gap: "18px", marginBottom: "18px" }}>
        {[...Array(6)].map((_, i) => (
          <div key={i} className="metric" style={{ minHeight: 120 }}>
            <div className="loading-pulse loading-pulse-short" style={{ width: 100, marginBottom: 12 }} />
            <div className="loading-pulse loading-pulse-medium" style={{ width: 60 }} />
          </div>
        ))}
      </div>

      <SkeletonPanel height={320} />
    </main>
  );
}

async function DashboardContent() {
  const dashboard = await getDashboard();
  const summary = dashboard.summary;
  const recentExecutions = dashboard.recent_executions.slice(0, 5);

  const metrics = [
    { label: "Active Skills", value: summary.active_skills },
    { label: "Published Versions", value: summary.published_versions },
    { label: "Running Executions", value: summary.running_executions },
    { label: "Waiting Approvals", value: summary.waiting_approvals },
    { label: "Success Rate", value: `${Math.round(summary.success_rate * 100)}%` },
    { label: "Avg Duration", value: formatDuration(summary.avg_duration_seconds) },
  ];

  return (
    <main className="shell">
      <section className="hero-page hero-compact">
        <p className="eyebrow">Observability</p>
        <h1>Skill hub at a glance.</h1>
        <p className="lede">
          Summary metrics and recent activity across your governed skill ecosystem.
        </p>
      </section>

      <div style={{ display: "grid", gridTemplateColumns: "repeat(3, 1fr)", gap: "18px", marginBottom: "18px" }}>
        {metrics.map((metric) => (
          <article className="metric" key={metric.label}>
            <p>{metric.label}</p>
            <strong>{metric.value}</strong>
          </article>
        ))}
      </div>

      <section className="panel">
        <div className="panel-header">
          <p className="section-kicker">Activity</p>
          <h2>Recent executions</h2>
        </div>
        <div className="list">
          {recentExecutions.map((execution) => (
            <article className="list-card" key={execution.id}>
              <div>
                <h3>{execution.skill_name}</h3>
                <p>
                  {execution.current_step_id} · started {new Date(execution.started_at).toLocaleString()}
                </p>
                <a
                  href={`/executions/${execution.id}/traces`}
                  className="trace-link"
                >
                  View trace →
                </a>
              </div>
              <span className={`badge badge-${execution.status}`}>{execution.status}</span>
            </article>
          ))}
        </div>
      </section>
    </main>
  );
}

export default function DashboardPage() {
  return (
    <Suspense fallback={<SkeletonDashboardPage />}>
      <DashboardContent />
    </Suspense>
  );
}
