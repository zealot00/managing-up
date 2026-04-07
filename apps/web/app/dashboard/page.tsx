import { Suspense } from "react";
import { getTranslations } from "next-intl/server";
import { getDashboard } from "../lib/api";
import { Package, Play, ClipboardCheck, CheckCircle } from "lucide-react";
import { EmptyState } from "../components/layout/EmptyState";

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
  const t = await getTranslations("dashboard");
  const dashboard = await getDashboard();
  const summary = dashboard.summary;
  const recentExecutions = dashboard.recent_executions.slice(0, 6);

  const metrics = [
    { label: t("totalSkills"), value: summary.active_skills, icon: <Package size={20} aria-hidden="true" /> },
    { label: t("activeExecutions"), value: summary.running_executions, icon: <Play size={20} aria-hidden="true" /> },
    { label: t("pendingApprovals"), value: summary.waiting_approvals, icon: <ClipboardCheck size={20} aria-hidden="true" /> },
    { label: t("avgSkillScore"), value: `${Math.round(summary.success_rate * 100)}%`, icon: <CheckCircle size={20} aria-hidden="true" /> },
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
          <h2 className="dashboard-section-title">{t("recentActivity")}</h2>
        </div>
        <div className="dashboard-list">
          {recentExecutions.length === 0 ? (
            <EmptyState title={t("noExecutions", { namespace: "executions" })} />
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
                    {t("executions:viewTrace")}
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
