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
    <>
      <div className="dashboard-stats">
        {[...Array(6)].map((_, i) => (
          <div key={i} className="dashboard-stat-card">
            <div className="loading-pulse loading-pulse-short" style={{ width: 80, marginBottom: 8 }} />
            <div className="loading-pulse" style={{ width: 60, height: 32 }} />
          </div>
        ))}
      </div>

      <div className="dashboard-section">
        <div className="loading-pulse loading-pulse-medium" style={{ marginBottom: 16 }} />
        <div className="skeleton-grid">
          {[1, 2, 3, 4, 5].map((i) => (
            <div key={i} className="skeleton-card" />
          ))}
        </div>
      </div>
    </>
  );
}

async function DashboardContent() {
  const dashboard = await getDashboard();
  const summary = dashboard.summary;
  const recentExecutions = dashboard.recent_executions.slice(0, 6);

  const metrics = [
    { label: "Active Skills", value: summary.active_skills, icon: "◉" },
    { label: "Published Versions", value: summary.published_versions, icon: "◎" },
    { label: "Running Executions", value: summary.running_executions, icon: "▸" },
    { label: "Waiting Approvals", value: summary.waiting_approvals, icon: "◐" },
    { label: "Success Rate", value: `${Math.round(summary.success_rate * 100)}%`, icon: "✓" },
    { label: "Avg Duration", value: formatDuration(summary.avg_duration_seconds), icon: "⏱" },
  ];

  return (
    <>
      <div className="dashboard-stats">
        {metrics.map((metric) => (
          <article className="dashboard-stat-card" key={metric.label}>
            <div className="dashboard-stat-icon">{metric.icon}</div>
            <div className="dashboard-stat-value">{metric.value}</div>
            <div className="dashboard-stat-label">{metric.label}</div>
          </article>
        ))}
      </div>

      <section className="dashboard-section">
        <div className="dashboard-section-header">
          <h2 className="dashboard-section-title">Recent executions</h2>
        </div>
        <div className="dashboard-list">
          {recentExecutions.length === 0 ? (
            <p className="empty-note">No executions yet</p>
          ) : (
            recentExecutions.map((execution) => (
              <article className="dashboard-list-item" key={execution.id}>
                <div className="dashboard-list-main">
                  <h3 className="dashboard-list-title">{execution.skill_name}</h3>
                  <p className="dashboard-list-meta">
                    {execution.current_step_id} · started {new Date(execution.started_at).toLocaleString()}
                  </p>
                </div>
                <div className="dashboard-list-actions">
                  <a href={`/executions/${execution.id}/traces`} className="dashboard-list-link">
                    View trace →
                  </a>
                  <span className={`badge badge-${execution.status}`}>{execution.status}</span>
                </div>
              </article>
            ))
          )}
        </div>
      </section>
    </>
  );
}

export default function DashboardPage() {
  return (
    <Suspense fallback={<SkeletonDashboardPage />}>
      <DashboardContent />
    </Suspense>
  );
}
