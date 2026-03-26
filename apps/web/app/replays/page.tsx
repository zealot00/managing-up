import { Suspense } from "react";
import { getReplaySnapshots } from "../lib/api";
import type { ReplaySnapshot } from "../lib/api";

function SkeletonReplays() {
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

function ReplayCard({ snap }: { snap: ReplaySnapshot }) {
  return (
    <article className="card">
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", gap: 12 }}>
        <div>
          <h2 style={{ margin: "0 0 8px", fontSize: "1.1rem" }}>Snapshot {snap.id.slice(0, 12)}</h2>
          <p style={{ margin: 0, color: "var(--muted)", fontSize: "0.82rem" }}>
            Execution: {snap.execution_id.slice(0, 16)}...
          </p>
        </div>
        <span className="badge badge-muted">
          Step {snap.step_index}
        </span>
      </div>
      <div className="detail-grid" style={{ marginTop: 16 }}>
        <div className="detail-row">
          <span className="detail-label">Skill ID</span>
          <span className="detail-value" style={{ fontSize: "0.82rem" }}>{snap.skill_id}</span>
        </div>
        <div className="detail-row">
          <span className="detail-label">Version</span>
          <span className="detail-value">{snap.skill_version}</span>
        </div>
        <div className="detail-row">
          <span className="detail-label">Seed</span>
          <span className="detail-value" style={{ fontFamily: "monospace", fontSize: "0.82rem" }}>
            {snap.deterministic_seed}
          </span>
        </div>
      </div>
      <div style={{ marginTop: 12, paddingTop: 12, borderTop: "1px solid var(--line)", fontSize: "0.82rem", color: "var(--muted)" }}>
        Created: {new Date(snap.created_at).toLocaleString()}
      </div>
    </article>
  );
}

async function ReplaysContent() {
  let snapshots: { items: ReplaySnapshot[] } | null = null;

  try {
    snapshots = await getReplaySnapshots();
  } catch {
    snapshots = null;
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

      <section className="hero-page hero-compact">
        <p className="eyebrow">Replay Layer</p>
        <h1>Replay Snapshots</h1>
        <p className="lede">
          Deterministic execution snapshots for replaying agent behavior.
          Each snapshot captures state, input seeds, and deterministic RNG seeds.
        </p>
      </section>

      <section aria-label="Snapshot list">
        {(snapshots?.items ?? []).length > 0 ? (
          <div className="eval-grid">
            {snapshots?.items.map((snap) => (
              <ReplayCard key={snap.id} snap={snap} />
            ))}
          </div>
        ) : (
          <article className="panel" style={{ marginTop: 24 }}>
            <div className="panel-header">
              <h2>No snapshots yet</h2>
            </div>
            <p style={{ color: "var(--muted)", marginTop: 12 }}>
              Replay snapshots capture execution state for deterministic replay.
              They are created when enabled during skill execution.
            </p>
          </article>
        )}
      </section>
    </main>
  );
}

export default function ReplaysPage() {
  return (
    <Suspense fallback={<SkeletonReplays />}>
      <ReplaysContent />
    </Suspense>
  );
}
