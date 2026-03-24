import { Suspense } from "react";
import { getDashboard, getSkills } from "./lib/api";
import type { DashboardData, Skill } from "./lib/api";
import { SkeletonPanel } from "./components/SkeletonPanel";

const platformModules = [
  { title: "Skill Registry", desc: "Version-controlled SOPs as executable skills with approval workflows." },
  { title: "Execution Engine", desc: "Stateful runtime with step tracking and approval checkpoints." },
  { title: "Task Evaluation", desc: "Reusable tasks with pluggable metrics for agent benchmarking." },
  { title: "Trace & Replay", desc: "Full execution traces and deterministic replay for debugging." },
];

const fallbackMetrics = [
  { label: "Active skills", value: "3" },
  { label: "Published versions", value: "3" },
  { label: "Running executions", value: "1" },
  { label: "Waiting approvals", value: "1" },
];

function SkeletonDashboard() {
  return (
    <main className="shell">
      <section className="toprail" aria-label="Platform posture">
        <div className="loading-pulse" style={{ width: 180, height: 44, borderRadius: 999 }} />
        <div className="loading-pulse" style={{ width: 200, height: 44, borderRadius: 999 }} />
        <div className="loading-pulse" style={{ width: 190, height: 44, borderRadius: 999 }} />
      </section>

      <section className="hero">
        <p className="eyebrow">Enterprise AI Quality Infrastructure</p>
        <div className="loading-pulse" style={{ width: 560, height: 48, marginBottom: 16 }} />
        <div className="loading-pulse loading-pulse-long" style={{ marginBottom: 8 }} />
        <div className="loading-pulse loading-pulse-medium" />
      </section>

      <section className="stats" aria-label="Platform overview">
        {[1, 2, 3, 4].map((i) => (
          <article className="metric" key={i}>
            <div className="loading-pulse loading-pulse-short" style={{ marginBottom: 12 }} />
            <div className="loading-pulse" style={{ width: 60, height: 32 }} />
          </article>
        ))}
      </section>

      <section className="panel-grid">
        <SkeletonPanel height={320} />
        <SkeletonPanel height={320} />
      </section>

      <section className="grid" aria-label="Platform modules">
        {[1, 2, 3, 4].map((i) => (
          <article className="card" key={i}>
            <div className="loading-pulse loading-pulse-short" style={{ marginBottom: 12 }} />
            <div className="loading-pulse loading-pulse-medium" />
          </article>
        ))}
      </section>
    </main>
  );
}

async function DashboardContent() {
  let dashboard: DashboardData | null = null;
  let skills: { items: Skill[] } | null = null;

  try {
    [dashboard, skills] = await Promise.all([getDashboard(), getSkills()]);
  } catch {
    dashboard = null;
    skills = null;
  }

  const metrics = dashboard
    ? [
        { label: "Active skills", value: String(dashboard.summary.active_skills) },
        { label: "Published versions", value: String(dashboard.summary.published_versions) },
        { label: "Running executions", value: String(dashboard.summary.running_executions) },
        { label: "Waiting approvals", value: String(dashboard.summary.waiting_approvals) },
      ]
    : fallbackMetrics;

  return (
    <main className="shell">
      <section className="toprail" aria-label="Platform posture">
        <div className="toprail-chip">Governed execution runtime</div>
        <div className="toprail-chip">Audit-ready operations</div>
        <div className="toprail-chip">Human approvals in loop</div>
        <a className="toprail-link" href="/skills">
          Registry
        </a>
        <a className="toprail-link" href="/executions">
          Executions
        </a>
        <a className="toprail-link" href="/approvals">
          Approvals
        </a>
        <a className="toprail-link" href="/tasks">
          Tasks
        </a>
        <a className="toprail-link" href="/evaluations">
          Evaluations
        </a>
      </section>

      <section className="hero">
        <p className="eyebrow">Enterprise AI Quality Infrastructure</p>
        <h1>Convert SOPs into executable skills.</h1>
        <p className="lede">
          Registry, execute, and observe AI agent skills with enterprise-grade controls.
          Full audit trails, human approvals, and task evaluation built in.
        </p>
      </section>

      <section className="stats" aria-label="Platform overview">
        {metrics.map((metric) => (
          <article className="metric" key={metric.label}>
            <p>{metric.label}</p>
            <strong>{metric.value}</strong>
          </article>
        ))}
      </section>

      <section className="panel-grid">
        <article className="panel">
          <div className="panel-header">
            <p className="section-kicker">Skill Registry</p>
            <h2>Published skills</h2>
          </div>
          <div className="list">
            {(skills?.items ?? []).length > 0 ? (
              skills?.items.map((skill) => (
                <article className="list-card" key={skill.id}>
                  <div>
                    <h3>{skill.name}</h3>
                    <p>
                      {skill.owner_team} · {skill.risk_level} risk
                    </p>
                  </div>
                  <span className={`badge badge-${skill.status}`}>{skill.status}</span>
                </article>
              ))
            ) : (
              <article className="list-card">
                <div>
                  <h3>Registry data unavailable</h3>
                  <p>Start the Go API on port 8080 to populate live skill data.</p>
                </div>
                <span className="badge badge-muted">offline</span>
              </article>
            )}
          </div>
        </article>

        <article className="panel">
          <div className="panel-header">
            <p className="section-kicker">Execution Timeline</p>
            <h2>Recent runs</h2>
          </div>
          <div className="list">
            {(dashboard?.recent_executions ?? []).length > 0 ? (
              dashboard?.recent_executions.map((execution) => (
                <article className="list-card" key={execution.id}>
                  <div>
                    <h3>{execution.skill_name}</h3>
                    <p>{execution.current_step_id}</p>
                  </div>
                  <span className={`badge badge-${execution.status}`}>{execution.status}</span>
                </article>
              ))
            ) : (
              <article className="list-card">
                <div>
                  <h3>No live execution feed</h3>
                  <p>Dashboard data appears when the backend dashboard endpoint is reachable.</p>
                </div>
                <span className="badge badge-muted">standby</span>
              </article>
            )}
          </div>
        </article>
      </section>

      <section className="grid" aria-label="Platform modules">
        {platformModules.map((mod) => (
          <article className="card" key={mod.title}>
            <h2>{mod.title}</h2>
            <p>{mod.desc}</p>
          </article>
        ))}
      </section>
    </main>
  );
}

export default function HomePage() {
  return (
    <Suspense fallback={<SkeletonDashboard />}>
      <DashboardContent />
    </Suspense>
  );
}
