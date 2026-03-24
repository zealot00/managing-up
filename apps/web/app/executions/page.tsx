import { Suspense } from "react";
import { getExecutions, getSkills } from "../lib/api";
import TriggerExecutionForm from "../components/TriggerExecutionForm";
import { SkeletonPanel } from "../components/SkeletonPanel";

function SkeletonExecutionsPage() {
  return (
    <main className="shell">
      <section className="hero hero-compact">
        <p className="eyebrow">Execution Timeline</p>
        <h1>Operational runs across governed skills.</h1>
        <p className="lede">
          Monitor live and completed runs, current step progression, and operator-triggered events.
        </p>
      </section>

      <div className="form-panel">
        <div className="loading-pulse loading-pulse-short" style={{ marginBottom: 12 }} />
        <div className="form-fields">
          <div className="loading-pulse loading-pulse-medium" />
          <div className="loading-pulse loading-pulse-medium" />
          <div className="loading-pulse loading-pulse-long" />
        </div>
        <div className="loading-pulse" style={{ width: 160, height: 44, borderRadius: 999 }} />
      </div>

      <SkeletonPanel height={320} />
    </main>
  );
}

async function ExecutionsContent() {
  const [executions, skills] = await Promise.all([getExecutions(), getSkills()]);

  return (
    <main className="shell">
      <section className="hero hero-compact">
        <p className="eyebrow">Execution Timeline</p>
        <h1>Operational runs across governed skills.</h1>
        <p className="lede">
          Monitor live and completed runs, current step progression, and operator-triggered events.
        </p>
      </section>

      <TriggerExecutionForm skills={skills.items} />

      <section className="panel">
        <div className="panel-header">
          <p className="section-kicker">Runs</p>
          <h2>Execution queue</h2>
        </div>
        <div className="list">
          {executions.items.map((execution) => (
            <article className="list-card" key={execution.id}>
              <div>
                <h3>{execution.skill_name}</h3>
                <p>
                  {execution.current_step_id} · triggered by {execution.triggered_by}
                </p>
                <a
                  href={`/executions/${execution.id}/traces`}
                  style={{ fontSize: "0.78rem", color: "var(--primary)", marginTop: 4, display: "inline-block" }}
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

export default function ExecutionsPage() {
  return (
    <Suspense fallback={<SkeletonExecutionsPage />}>
      <ExecutionsContent />
    </Suspense>
  );
}
