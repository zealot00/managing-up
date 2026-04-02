import { Suspense } from "react";
import { getDashboard } from "../lib/api";

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
      <header className="hero-page hero-compact">
        <p className="eyebrow">Observability</p>
        <h1>Skill hub at a glance.</h1>
        <p className="lede">
          Summary metrics and recent activity across your governed skill ecosystem.
        </p>
      </header>

      <div className="stats">
        {[...Array(6)].map((_, i) => (
          <div key={i} className="metric-card">
            <div className="loading-pulse loading-pulse-short" style={{ width: 80, marginBottom: 8 }} />
            <div className="loading-pulse" style={{ width: 60, height: 32 }} />
          </div>
        ))}
      </div>

      <div className="panel">
        <div className="loading-pulse loading-pulse-medium" style={{ marginBottom: 16 }} />
        <div className="skeleton-grid">
          {[1, 2, 3, 4, 5].map((i) => (
            <div key={i} className="skeleton-card" />
          ))}
        </div>
      </div>
    </main>
  );
}

async function DashboardContent() {
  const dashboard = await getDashboard();
  const summary = dashboard.summary;
  const recentExecutions = dashboard.recent_executions.slice(0, 6);

  const metrics = [
    { label: "Active Skills", value: summary.active_skills, icon: "◉" },
    { label: "Published Versions", value: summary.published_versions, icon: "◎" },
    { label: "Running Executions", value: summary.running_executions, icon: "▶" },
    { label: "Waiting Approvals", value: summary.waiting_approvals, icon: "◐" },
    { label: "Success Rate", value: `${Math.round(summary.success_rate * 100)}%`, icon: "✓" },
    { label: "Avg Duration", value: formatDuration(summary.avg_duration_seconds), icon: "⏱" },
  ];

  return (
    <main className="shell">
      <header className="hero-page hero-compact">
        <p className="eyebrow">Observability</p>
        <h1>Skill hub at a glance.</h1>
        <p className="lede">
          Summary metrics and recent activity across your governed skill ecosystem.
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

      <section className="panel">
        <div className="panel-header">
          <p className="section-kicker">Activity</p>
          <h2 className="panel-title">Recent executions</h2>
        </div>
        <div className="list">
          {recentExecutions.length === 0 ? (
            <p className="empty-note">No executions yet</p>
          ) : (
            recentExecutions.map((execution) => (
              <article className="list-card" key={execution.id}>
                <div className="list-card-main">
                  <h3 className="list-card-title">{execution.skill_name}</h3>
                  <p className="list-card-meta">
                    {execution.current_step_id} · started {new Date(execution.started_at).toLocaleString()}
                  </p>
                </div>
                <div className="list-card-actions">
                  <a href={`/executions/${execution.id}/traces`} className="trace-link">
                    View trace →
                  </a>
                  <span className={`badge badge-${execution.status}`}>{execution.status}</span>
                </div>
              </article>
            ))
          )}
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
