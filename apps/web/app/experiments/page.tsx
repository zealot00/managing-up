import { Suspense } from "react";
import { getExperiments } from "../lib/api";
import type { Experiment } from "../lib/api";

function SkeletonExperiments() {
  return (
    <main className="shell">
      <header className="hero-page hero-compact">
        <p className="eyebrow">Experiment DB</p>
        <h1>Experiments</h1>
        <p className="lede">
          Compare agent and skill performance across multiple task executions.
        </p>
      </header>
      <div className="skeleton-grid">
        {[1, 2, 3].map((i) => (
          <div key={i} className="skeleton-card" />
        ))}
      </div>
    </main>
  );
}

function ExperimentCard({ exp }: { exp: Experiment }) {
  return (
    <article className="eval-card">
      <div className="eval-card-header">
        <div>
          <h3 className="eval-card-title">{exp.name}</h3>
          <p className="eval-card-meta">{exp.description || "No description"}</p>
        </div>
        <span className={`badge badge-${exp.status === "completed" ? "succeeded" : exp.status === "running" ? "running" : "muted"}`}>
          {exp.status}
        </span>
      </div>
      <div className="tags">
        <span className="tag">{exp.task_ids.length} tasks</span>
        <span className="tag">{exp.agent_ids.length} agents</span>
      </div>
      <div className="eval-card-footer">
        <span>Created: {new Date(exp.created_at).toLocaleString()}</span>
      </div>
    </article>
  );
}

async function ExperimentsContent() {
  let experiments: { items: Experiment[] } | null = null;

  try {
    experiments = await getExperiments();
  } catch {
    experiments = null;
  }

  return (
    <main className="shell">
      <header className="hero-page hero-compact">
        <p className="eyebrow">Experiment DB</p>
        <h1>Experiments</h1>
        <p className="lede">
          Compare agent and skill performance across multiple task executions.
          Each experiment defines a set of tasks and agents to evaluate.
        </p>
      </header>

      <section aria-label="Experiment list">
        {(experiments?.items ?? []).length > 0 ? (
          <div className="eval-grid">
            {experiments?.items.map((exp) => (
              <ExperimentCard key={exp.id} exp={exp} />
            ))}
          </div>
        ) : (
          <div className="empty-state">
            <div className="empty-state-icon">◎</div>
            <h3 className="empty-state-title">No experiments yet</h3>
            <p className="empty-state-description">
              Experiments will appear here once created via POST /api/v1/experiments.
            </p>
          </div>
        )}
      </section>
    </main>
  );
}

export default function ExperimentsPage() {
  return (
    <Suspense fallback={<SkeletonExperiments />}>
      <ExperimentsContent />
    </Suspense>
  );
}
