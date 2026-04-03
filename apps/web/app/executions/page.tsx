import { Suspense } from "react";
import { getExecutions, getSkills } from "../lib/api";
import TriggerExecutionForm from "../components/TriggerExecutionForm";

function SkeletonExecutionsPage() {
  return (
    <main className="shell">
      <header className="hero-page hero-compact">
        <p className="eyebrow">Execution Timeline</p>
        <h1>Operational runs across governed skills.</h1>
        <p className="lede">
          Monitor live and completed runs, current step progression, and operator-triggered events.
        </p>
      </header>

      <div className="form-panel">
        <div className="loading-pulse loading-pulse-short" style={{ marginBottom: 16 }} />
        <div className="form-fields">
          <div className="loading-pulse loading-pulse-medium" />
          <div className="loading-pulse loading-pulse-medium" />
          <div className="loading-pulse loading-pulse-long" />
        </div>
      </div>

      <div className="panel">
        <div className="loading-pulse loading-pulse-medium" style={{ marginBottom: 16 }} />
        <div className="skeleton-grid">
          {[1, 2, 3, 4, 5].map((i) => <div key={i} className="skeleton-card" />)}
        </div>
      </div>
    </main>
  );
}

async function ExecutionsContent() {
  const [executions, skills] = await Promise.all([getExecutions(), getSkills()]);

  return (
    <main className="shell">
      <header className="hero-page hero-compact">
        <p className="eyebrow">Execution Timeline</p>
        <h1>Operational runs across governed skills.</h1>
        <p className="lede">
          Monitor live and completed runs, current step progression, and operator-triggered events.
        </p>
      </header>

      <TriggerExecutionForm skills={skills.items} />

      <section className="panel">
        <div className="panel-header">
          <p className="section-kicker">Runs</p>
          <h2 className="panel-title">Execution queue</h2>
        </div>
        <div className="list">
          {executions.items.length === 0 ? (
            <p className="empty-note">No executions in queue</p>
          ) : (
            executions.items.map((execution) => (
              <article className="list-card" key={execution.id}>
                <div className="list-card-main">
                  <h3 className="list-card-title">{execution.skill_name}</h3>
                  <p className="list-card-meta">
                    {execution.current_step_id} · triggered by {execution.triggered_by}
                  </p>
                </div>
                <div className="list-card-actions">
                  <a href={`/executions/${execution.id}`} className="trace-link">
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

export default function ExecutionsPage() {
  return (
    <Suspense fallback={<SkeletonExecutionsPage />}>
      <ExecutionsContent />
    </Suspense>
  );
}
