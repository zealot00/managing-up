import { Suspense } from "react";
import { getExperiments } from "../lib/api";
import type { Experiment } from "../lib/api";

function SkeletonExperiments() {
  return (
    <main className="shell">
      <section className="toprail">
        <div className="loading-pulse" style={{ width: 180, height: 44, borderRadius: 999 }} />
      </section>
      <div className="loading-pulse loading-pulse-medium" style={{ marginBottom: 8 }} />
      <div className="skeleton-grid">
        {[1, 2, 3].map((i) => (
          <div className="skeleton-card" key={i} />
        ))}
      </div>
    </main>
  );
}

function ExperimentCard({ exp }: { exp: Experiment }) {
  return (
    <article className="card">
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", gap: 12 }}>
        <div>
          <h2 style={{ margin: "0 0 8px", fontSize: "1.2rem" }}>{exp.name}</h2>
          <p style={{ margin: 0, color: "var(--muted)", fontSize: "0.9rem" }}>{exp.description || "No description"}</p>
        </div>
        <span className={`badge badge-${exp.status === "completed" ? "succeeded" : exp.status === "running" ? "running" : "muted"}`}>
          {exp.status}
        </span>
      </div>
      <div style={{ marginTop: 16, display: "flex", gap: 8, flexWrap: "wrap" }}>
        <span style={{ padding: "4px 10px", borderRadius: 999, background: "rgba(36, 49, 64, 0.06)", fontSize: "0.78rem", color: "var(--ink)" }}>
          {exp.task_ids.length} tasks
        </span>
        <span style={{ padding: "4px 10px", borderRadius: 999, background: "rgba(36, 49, 64, 0.06)", fontSize: "0.78rem", color: "var(--ink)" }}>
          {exp.agent_ids.length} agents
        </span>
      </div>
      <div style={{ marginTop: 12, paddingTop: 12, borderTop: "1px solid var(--line)", fontSize: "0.82rem", color: "var(--muted)" }}>
        Created: {new Date(exp.created_at).toLocaleString()}
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
      <section className="toprail">
        <a className="toprail-link" href="/">
          Dashboard
        </a>
        <a className="toprail-link" href="/tasks">
          Tasks
        </a>
        <a className="toprail-link" href="/evaluations">
          Evaluations
        </a>
        <a className="toprail-link" href="/experiments">
          Experiments
        </a>
        <a className="toprail-link" href="/replays">
          Replays
        </a>
      </section>

      <section className="hero-compact">
        <p className="eyebrow">Experiment DB</p>
        <h1>Experiments</h1>
        <p className="lede">
          Compare agent and skill performance across multiple task executions.
          Each experiment defines a set of tasks and agents to evaluate.
        </p>
      </section>

      <section aria-label="Experiment list">
        {(experiments?.items ?? []).length > 0 ? (
          <div className="eval-grid">
            {experiments?.items.map((exp) => (
              <ExperimentCard key={exp.id} exp={exp} />
            ))}
          </div>
        ) : (
          <article className="panel" style={{ marginTop: 24 }}>
            <div className="panel-header">
              <h2>No experiments yet</h2>
            </div>
            <p style={{ color: "var(--muted)", marginTop: 12 }}>
              Experiments will appear here once created via POST /api/v1/experiments.
            </p>
          </article>
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
